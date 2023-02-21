package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/bbt-t/ya-go-d/internal/entity"

	"golang.org/x/crypto/bcrypt"
)

func (s *dbStorage) NewUser(ctx context.Context, user entity.User) (int, error) {
	/*
		Creating and insert (db) new user.
	*/
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.WaitingTime)
	defer cancel()

	if _, err := s.GetUser(ctx, entity.SearchByLogin, user.Login); err == nil {
		return 0, ErrExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Login+user.Password), 0)
	if err != nil {
		log.Println("Failed generate hash from password:", err)
		return 0, err
	}

	if _, err = s.DB.ExecContext(
		ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2)",
		user.Login,
		hash,
	); err != nil {
		log.Println("Failed added new user to DB:", err)
		return 0, err
	}

	if err = s.DB.QueryRowContext(
		ctx,
		"SELECT id FROM users WHERE login = $1",
		user.Login,
	).Scan(&user.ID); err != nil {
		log.Println("Failed get user ID:", err)
		return 0, err
	}
	return user.ID, nil
}

func (s *dbStorage) GetUser(ctx context.Context, search, value string) (entity.User, error) {
	var (
		row  *sql.Row
		user entity.User
	)

	ctx, cancel := context.WithTimeout(ctx, s.Cfg.WaitingTime)
	defer cancel()

	switch search {
	case entity.SearchByID:
		row = s.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", value)
	case entity.SearchByLogin:
		row = s.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE login = $1", value)
	default:
		log.Fatalln("Failed search user by type")
		return user, errors.New("received wrong search type")
	}

	switch err := row.Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Balance,
		&user.Withdrawn,
	); err {
	case sql.ErrNoRows:
		return user, ErrNotFound
	case nil:
		return user, nil
	default:
		log.Println("Failed get user:", err)
		return user, err
	}
}

func (s *dbStorage) Withdraw(ctx context.Context, user entity.User, wd entity.Withdraw) error {
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.WaitingTime)
	defer cancel()

	result, err := s.DB.ExecContext(
		ctx,
		"UPDATE users SET balance = balance - $1, withdrawn = withdrawn + $1 WHERE id = $2 AND balance >= $1",
		wd.Sum,
		user.ID,
	)
	if err != nil {
		log.Println("Failed withdraw:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Failed get affected rows")
		return err
	}
	if rowsAffected == 0 {
		return ErrNoEnoughBalance
	}

	if _, err = s.DB.ExecContext(
		ctx,
		"INSERT INTO withdrawals (user_id, num, amount) VALUES ($1, $2, $3)",
		user.ID,
		wd.Order,
		wd.Sum,
	); err != nil {
		log.Println("Failed insert withdrawal into withdrawals:", err)
		return err
	}
	return nil
}

func (s *dbStorage) WithdrawAll(ctx context.Context, user entity.User) ([]entity.Withdraw, error) {
	var withdrawals []entity.Withdraw

	ctx, cancel := context.WithTimeout(ctx, s.Cfg.WaitingTime)
	defer cancel()

	rows, err := s.DB.QueryContext(
		ctx,
		"SELECT num, amount, processed FROM withdrawals WHERE user_id = $1 ORDER BY processed DESC",
		user.ID,
	)

	if err != nil {
		log.Println("Can't get withdrawals history from DB:", err)
		return withdrawals, err
	}

	for rows.Next() {
		withdrawal := entity.Withdraw{}
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.Time)
		if err != nil {
			log.Println("Error while scanning rows:", err)
			return withdrawals, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		log.Println("Rows error:", err)
		return withdrawals, err
	}
	return withdrawals, nil
}

func (s *dbStorage) AddOrder(ctx context.Context, order entity.Order) error {
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.WaitingTime)
	defer cancel()

	row := s.DB.QueryRowContext(
		ctx,
		"SELECT user_id FROM orders WHERE num = $1 LIMIT 1",
		order.Number,
	)

	orderDB := entity.Order{}
	err := row.Scan(&orderDB.UserID)
	if err == nil {
		if orderDB.UserID == order.UserID {
			return ErrNumAlreadyLoaded
		}
		return ErrWrongUser
	}

	_, err = s.DB.ExecContext(
		ctx,
		"INSERT INTO orders (user_id, num) VALUES ($1, $2)",
		order.UserID,
		order.Number,
	)

	if err != nil {
		log.Println("Failed insert new order into orders:", err)
		return err
	}

	err = s.Queue.PushBack(order)
	if err != nil {
		log.Println("Failed push order to queue")
		return err
	}
	return nil
}

func (s *dbStorage) OrdersAll(ctx context.Context, user entity.User) ([]entity.Order, error) {
	var orders []entity.Order

	ctx, cancel := context.WithTimeout(ctx, s.Cfg.WaitingTime)
	defer cancel()

	rows, err := s.DB.QueryContext(
		ctx,
		"SELECT num, status, accrual, uploaded FROM orders WHERE user_id = $1 ORDER BY uploaded DESC ",
		user.ID,
	)

	if err != nil {
		log.Println("Can't get withdrawals history from DB:", err)
		return orders, err
	}

	for rows.Next() {
		order := entity.Order{}
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.EventTime)
		if err != nil {
			log.Println("Error while scanning rows:", err)
			return orders, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		log.Println("Rows error:", err)
		return orders, err
	}
	return orders, nil
}

func (s *dbStorage) GetOrderForUpdate() (entity.Order, error) {
	return s.Queue.GetOrder()
}

func (s *dbStorage) GetOrdersForUpdate(ctx context.Context) ([]entity.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.WaitingTime)
	defer cancel()

	var orders []entity.Order
	query := `SELECT user_id, num, status FROM orders 
              WHERE (stat = 'NEW' OR status = 'PROCESSING') AND  
              uploaded < $1 ORDER BY uploaded ASC LIMIT 10`
	rows, err := s.DB.QueryContext(
		ctx,
		query,
		s.StartTime,
	)

	if err != nil {
		log.Println("can't get orders from DB for update:", err)
		return orders, err
	}

	for rows.Next() {
		order := entity.Order{}
		err = rows.Scan(&order.UserID, &order.Number, &order.Status)
		if err != nil {
			log.Println("Rows error:", err)
			return orders, err
		}
		orders = append(orders, order)
	}

	err = rows.Err()
	if errors.Is(err, sql.ErrNoRows) || err == nil {
		return orders, nil
	}

	return orders, err
}

func (s *dbStorage) UpdateOrders(ctx context.Context, orders ...entity.Order) error {
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.WaitingTime)
	defer cancel()

	tx, err := s.DB.Begin()
	defer tx.Rollback()

	if err != nil {
		return err
	}

	stmtOrders, err := tx.PrepareContext(ctx, `UPDATE orders SET status = $1, accrual = $2 WHERE num = $3`)
	if err != nil {
		return err
	}
	defer stmtOrders.Close()

	stmtUsers, err := tx.PrepareContext(ctx, `UPDATE users SET balance = balance + $1 WHERE id = $2`)
	if err != nil {
		return err
	}
	defer stmtUsers.Close()

	for _, order := range orders {
		if _, err := stmtOrders.ExecContext(ctx, order.Status, order.Accrual, order.Number); err != nil {
			return err
		}

		if order.Accrual > 0 {
			if _, err := stmtUsers.ExecContext(ctx, order.Accrual, order.UserID); err != nil {
				return err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *dbStorage) Push(orders []entity.Order) error {
	return s.Queue.Push(orders)
}

func (s *dbStorage) PushBack(order entity.Order) error {
	return s.Queue.PushBack(order)
}
