package logging

import "log"

func init() {
	/*
		Logging settings.
	*/
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}
