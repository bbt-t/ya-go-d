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
	ctx, cancel := context.WithTimeout(ctx, s.cfg.WaitingTime)
	defer cancel()

	if _, err := s.GetUser(ctx, entity.SearchByLogin, user.Login); err == nil {
		return 0, ErrExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Login+user.Password), 0)
	if err != nil {
		log.Printf("Failed generate hash from password: %+v\n", err)
		return 0, err
	}

	if _, err = s.db.ExecContext(
		ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2)",
		user.Login,
		hash,
	); err != nil {
		log.Printf("Failed added new user to DB: %+v\n", err)
		return 0, err
	}

	if err = s.db.QueryRowContext(
		ctx,
		"SELECT id FROM users WHERE login = $1",
		user.Login,
	).Scan(&user.ID); err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (s *dbStorage) GetUser(ctx context.Context, search, value string) (entity.User, error) {
	var (
		row  *sql.Row
		user entity.User
	)

	ctx, cancel := context.WithTimeout(ctx, s.cfg.WaitingTime)
	defer cancel()

	switch search {
	case entity.SearchByID:
		row = s.db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", value)
	case entity.SearchByLogin:
		row = s.db.QueryRowContext(ctx, "SELECT * FROM users WHERE login = $1", value)
	default:
		return user, ErrSearchType
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
		return user, err
	}
}

func (s *dbStorage) Withdraw(ctx context.Context, user entity.User, wd entity.Withdraw) error {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.WaitingTime)
	defer cancel()

	result, err := s.db.ExecContext(
		ctx,
		"UPDATE users SET balance = balance - $1, withdrawn = withdrawn + $1 WHERE id = $2 AND balance >= $1",
		wd.Sum,
		user.ID,
	)
	if err != nil {
		log.Printf("Failed withdraw: %+v\n", err)
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

	if _, err = s.db.ExecContext(
		ctx,
		"INSERT INTO withdrawals (user_id, num, amount) VALUES ($1, $2, $3)",
		user.ID,
		wd.Order,
		wd.Sum,
	); err != nil {
		log.Printf("Failed insert withdrawal into withdrawals: %+v\n", err)
		return err
	}
	return nil
}

func (s *dbStorage) WithdrawAll(ctx context.Context, user entity.User) ([]entity.Withdraw, error) {
	var withdrawals []entity.Withdraw

	ctx, cancel := context.WithTimeout(ctx, s.cfg.WaitingTime)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT num, amount, processed FROM withdrawals WHERE user_id = $1 ORDER BY processed DESC",
		user.ID,
	)

	if err != nil {
		return withdrawals, err
	}

	for rows.Next() {
		withdrawal := entity.Withdraw{}
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.Time)
		if err != nil {
			return withdrawals, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return withdrawals, err
	}
	return withdrawals, nil
}

func (s *dbStorage) AddOrder(ctx context.Context, order entity.Order) error {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.WaitingTime)
	defer cancel()

	row := s.db.QueryRowContext(
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

	_, err = s.db.ExecContext(
		ctx,
		"INSERT INTO orders (user_id, num) VALUES ($1, $2)",
		order.UserID,
		order.Number,
	)

	if err != nil {
		log.Printf("Failed insert new order into orders: %+v\n", err)
		return err
	}

	err = s.queue.PushBack(order)
	if err != nil {
		return err
	}
	return nil
}

func (s *dbStorage) OrdersAll(ctx context.Context, user entity.User) ([]entity.Order, error) {
	var orders []entity.Order

	ctx, cancel := context.WithTimeout(ctx, s.cfg.WaitingTime)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT num, status, accrual, uploaded FROM orders WHERE user_id = $1 ORDER BY uploaded DESC ",
		user.ID,
	)

	if err != nil {
		return orders, err
	}

	for rows.Next() {
		order := entity.Order{}
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.EventTime)
		if err != nil {
			return orders, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return orders, err
	}
	return orders, nil
}

func (s *dbStorage) GetOrderForUpdate() (entity.Order, error) {
	return s.queue.GetOrder()
}

func (s *dbStorage) GetOrdersForUpdate(ctx context.Context) ([]entity.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.WaitingTime)
	defer cancel()

	var orders []entity.Order
	query := `SELECT user_id, num, status FROM orders 
              WHERE (stat = 'NEW' OR status = 'PROCESSING') AND  
              uploaded < $1 ORDER BY uploaded ASC LIMIT 10`
	rows, err := s.db.QueryContext(
		ctx,
		query,
		s.startTime,
	)

	if err != nil {
		return orders, err
	}

	for rows.Next() {
		order := entity.Order{}
		err = rows.Scan(&order.UserID, &order.Number, &order.Status)
		if err != nil {
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
	ctx, cancel := context.WithTimeout(ctx, s.cfg.WaitingTime)
	defer cancel()

	tx, err := s.db.Begin()
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
	return s.queue.Push(orders)
}

func (s *dbStorage) PushBack(order entity.Order) error {
	return s.queue.PushBack(order)
}
