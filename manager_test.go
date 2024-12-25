package session_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/enorith/session"
	"github.com/enorith/session/handlers"
	"github.com/enorith/supports/str"
)

type Sess struct {
	Name string
}

func TestManater(t *testing.T) {
	id := str.RandString(32)
	h := handlers.NewFileSessionHandler(".\\temp\\session")
	s := session.NewManager(h)
	cur := 6
	e := s.Start(id)
	if e != nil {
		fmt.Println(e)
		return
	}

	for i := 0; i < cur; i++ {
		go func(idx int) {
			var se Sess
			se.Name = fmt.Sprintf("foo %d", idx)
			s.Get(id).Set("se", se)
			//fmt.Println(se)
		}(i)
		go func(idx int) {
			var se Sess
			s.Get(id).Get("se", &se)
			fmt.Println(se)
		}(i)
	}
}

func TestSession(t *testing.T) {
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			testSess()
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkSession(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testSess()
	}
}

func testSess() {
	id := str.RandString(32)
	h := handlers.NewFileSessionHandler(".\\temp\\session")
	s := session.NewManager(h)
	cur := 6
	e := s.Start(id)
	if e != nil {
		fmt.Println(e)
		return
	}

	wg := sync.WaitGroup{}

	for i := 0; i < cur; i++ {
		wg.Add(1)
		go func(idx int) {
			var se Sess
			se.Name = fmt.Sprintf("foo %d", idx)
			s.Get(id).Set("se", se)
			//fmt.Println(se)
			wg.Done()
		}(i)
		wg.Add(1)

		go func(idx int) {
			var se Sess
			s.Get(id).Get("se", &se)
			//fmt.Println(se)
			wg.Done()
		}(i)
	}

	wg.Wait()
}
