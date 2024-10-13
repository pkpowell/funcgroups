package funcgroups

import (
	"fmt"
	"testing"
	"time"
)

func one() {
	fmt.Println("func one")
	time.Sleep(time.Second * 1)
	fmt.Println("func one done")
}
func two() {
	fmt.Println("func two")
	time.Sleep(time.Second * 2)
	fmt.Println("func two done")
}
func three() {
	fmt.Println("func three")
	time.Sleep(time.Second * 3)
	fmt.Println("func three done")
}
func four() {
	fmt.Println("func four")
	time.Sleep(time.Second * 4)
	fmt.Println("func four done")
}

func RunWait_test(t *testing.T) {
	t.Log("RunWait_test")
	RunWait([]Function{one, two, three, four}, &Options{
		Timeout: 5 * time.Second,
		// Ctx:     context.Background(),
		// Cancel:  cancel,
	})
}
