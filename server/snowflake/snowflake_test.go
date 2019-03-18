package snowflake

import "testing"

func TestNewFactory(t *testing.T) {
	_, err := NewFactory(-1)
	if err == nil {
		t.Fatal("Must to be failed, machineId = -1")
	}
	_, err = NewFactory(1024)
	if err == nil {
		t.Fatal("Must to be failed, machineId = -1")
	}
}

func TestFactory(t *testing.T) {
	f, err := NewFactory(100)
	if err != nil {
		t.Fatal(err)
	}
	var last int64
	for i := 0; i < 1000; i++ {
		id, err := f.Next()
		if err != nil {
			t.Fatal(err)
		}
		if id <= last {
			t.Fatalf("got id = %d, last = %d", id, last)
		}
		last = id
	}
}

func BenchmarkFactory(b *testing.B) {
	f, err := NewFactory(100)
	if err != nil {
		b.Fatal(err)
	}
	var last int64
	for i := 0; i < b.N; i++ {
		id, err := f.Next()
		if err != nil {
			b.Fatal(err)
		}
		if id <= last {
			b.Fatalf("got id = %d, last = %d", id, last)
		}
		last = id
	}
}

func BenchmarkFactoryParallel1(b *testing.B) {
	factories := make(chan Factory, 1)
	f, _ := NewFactory(100)
	factories <- f

	gen := func() (int64, error) {
		f := <-factories
		defer func() {
			factories <- f
		}()
		return f.Next()
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var last int64
			for i := 0; i < 100; i++ {
				id, err := gen()
				if err != nil {
					b.Fatal(err)
				}
				if id <= last {
					b.Fatalf("last = %d, id = %d", last, id)
				}
				last = id
			}
		}
	})
}

func BenchmarkFactoryParallel2(b *testing.B) {
	factories := make(chan Factory, 1)
	f, _ := NewFactory(100)
	factories <- f
	gen2 := func(count int) (ids []int64, err error) {
		f := <-factories
		defer func() {
			factories <- f
		}()
		var last int64
		for i := 0; i < 10; i++ {
			id, err := f.Next()
			if err != nil {
				b.Fatal(err)
			}
			if id <= last {
				b.Fatalf("last = %d, id = %d", last, id)
			}
			last = id
			// ids = append(ids, id)
		}
		return ids, err
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < 100; i++ {
				_, err := gen2(10)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}
