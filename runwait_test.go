package funcgroups

import (
	"fmt"
	"testing"
	"time"
)

func one(a any) {
	fmt.Println("func one")
	time.Sleep(time.Second * 1)
	fmt.Println("func one done")
}
func two(a any) {
	fmt.Println("func two")
	time.Sleep(time.Second * 2)
	fmt.Println("func two done")
}
func three(a any) {
	fmt.Println("func three")
	time.Sleep(time.Second * 3)
	fmt.Println("func three done")
}
func four(a any) {
	fmt.Println("func four")
	time.Sleep(time.Second * 4)
	fmt.Println("func four done")
}

func RunWait_test(t *testing.T) {
	t.Log("RunWait_test")
	RunWait([]Function{one, two, three, four}, t, &Options{
		Timeout: 5 * time.Second,
		// Ctx:     context.Background(),
		// Cancel:  cancel,
	})
}
