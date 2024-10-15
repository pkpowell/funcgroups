package funcgroups

import (
	"context"
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
func fifteen() {
	fmt.Println("func four")
	time.Sleep(time.Second * 4)
	fmt.Println("func four done")
}

var allFuncs = []Function{one, two, three, four, fifteen}

func TestRunWait(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		RunWait_test(t)
	})

	t.Run("Timeout scenario", func(t *testing.T) {
		start := time.Now()
		RunWait(allFuncs, &Options{
			Timeout: 3 * time.Second,
			Ctx:     context.Background(),
			Debug:   BoolPointer(true),
		})
		duration := time.Since(start)
		if duration > 3*time.Second+100*time.Millisecond {
			t.Errorf("RunWait didn't respect timeout. Took %v, expected around 3s", duration)
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(2 * time.Second)
			cancel()
		}()

		start := time.Now()
		RunWait(allFuncs, &Options{
			Timeout: 10 * time.Second,
			Ctx:     ctx,
			Debug:   BoolPointer(true),
		})
		duration := time.Since(start)
		if duration > 2*time.Second+100*time.Millisecond {
			t.Errorf("RunWait didn't respect context cancellation. Took %v, expected around 2s", duration)
		}
	})

	t.Run("Empty function list", func(t *testing.T) {
		RunWait([]Function{}, &Options{
			Timeout: 1 * time.Second,
			Ctx:     context.Background(),
			Debug:   BoolPointer(true),
		})
		// This test passes if it doesn't panic
	})

	t.Run("Nil options", func(t *testing.T) {
		RunWait([]Function{one, two}, nil)
		// This test passes if it doesn't panic and uses default options
	})
}

func RunWait_test(t *testing.T) {
	t.Log("RunWait_test")
	RunWait(allFuncs, &Options{
		Timeout: 2 * time.Second,
		Ctx:     context.Background(),
		Debug:   BoolPointer(true),
	})
}
