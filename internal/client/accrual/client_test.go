package accrual

import (
	"log"
	"testing"
)

func TestAccrualClient(t *testing.T) {
	client := NewAccrualClient(100, "localhost:8090")

	order, err := client.GetData("4739242")
	if err != nil {
		t.Errorf("ошибка возникла %v", err)
	}

	log.Println(order)
}
