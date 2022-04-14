package wrr

import (
	"fmt"
	"net/http"
	"testing"
)

func TestWrrLooper_selectTarget(t *testing.T) {
	w := &WrrLooper[string]{}
	w.defaultWeight = 100
	w.deltaWeight = 0
	w.SetTargets(&Target[string]{100, "10.1.87.70:8213"}, &Target[string]{100, "10.1.87.70:8212"}, &Target[string]{10, "10.1.87.70:8211"})
	for i := 0; i < 200; i++ {
		w.Call(func(t string) error {
			fmt.Println("call:", t)
			_, err := http.Get("http://" + t)
			return err
		})
	}

}
