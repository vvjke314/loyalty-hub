package accrual

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
	"go.uber.org/ratelimit"
)

var ErrTooFrequentRequests = errors.New("too many requests to outer service")
var ErrDataIsNotArrived = errors.New("no data about this order")

// структура которая отправляет запроса во внешний сервис
// ratelimit.Limiter использует leacky-bucket алгоритм, который распределяет равномерно отправку
// блокирует попытки запросов чтобы укладываться в rps
type AccrualClient struct {
	reqUrl    string
	rps       int
	client    *http.Client
	ratelimit ratelimit.Limiter
}

func NewAccrualClient(rps int, reqUrl string) *AccrualClient {
	return &AccrualClient{
		reqUrl:    reqUrl,
		rps:       rps,
		client:    &http.Client{},
		ratelimit: ratelimit.New(rps),
	}
}

// отправляем запросы на сервер
// запускать переодически чтобы обновлять информацию в базе данных
func (a *AccrualClient) GetData(orderNum string) (dto.AccrualServiceResponse, error) {
	a.ratelimit.Take()

	resp, err := a.client.Get("http://" + a.reqUrl + "/api/orders/" + orderNum)
	if err != nil {
		return dto.AccrualServiceResponse{}, err
	}

	var data dto.AccrualServiceResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	defer resp.Body.Close()
	if err != nil {
		return dto.AccrualServiceResponse{}, err
	}

	return data, nil
}

func (a *AccrualClient) GetRPS() int {
	return a.rps
}
