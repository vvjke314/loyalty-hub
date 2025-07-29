package repository

import "errors"

// ошибка если нет пользователей
var ErrNoUser = errors.New("no such user in db")