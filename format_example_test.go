package formatting

import (
	"fmt"
)

func ExampleFormat() {
	type A struct {
		VInt int
	}
	type B struct {
		VBool       bool
		VInt        int
		VInt32      int32
		VFloat64    float64
		VComplex128 complex128
		VString     string
		VA          A
		PA          *A
		PA2         *A
		Map         map[string]*A
		Map2        map[string]*A
		List        []interface{}
		List2       []interface{}
		F           func(format string, args ...interface{}) (int, error)
		F2          func(format string, args ...interface{}) (int, error)
		Ch          chan string
	}

	b := &B{
		VA:          A{1},
		PA:          &A{2},
		VInt:        3,
		VInt32:      4,
		VFloat64:    5.0,
		VComplex128: complex(6, 7),
		VString:     "8",
		Map:         map[string]*A{"a": &A{9}},
		Map2:        nil,
		List:        []interface{}{A{10}, 10},
		List2:       nil,
		F:           fmt.Printf,
		Ch:          make(chan string, 20),
	}
	b.Ch <- "hello"

	fmt.Println(Format(b))

	// Output:
	// &formatting.B{
	//     VBool: bool{false},
	//     VInt: int{3},
	//     VInt32: int32{4},
	//     VFloat64: float64{5},
	//     VComplex128: complex{(6+7i)},
	//     VString: string{"8"},
	//     VA: formatting.A{
	//         VInt: int{1},
	//     },
	//     PA: &formatting.A{
	//         VInt: int{2},
	//     },
	//     PA2: (*formatting.A)(<nil>),
	//     Map: map{
	//         string{"a"}: &formatting.A{
	//             VInt: int{9},
	//         },
	//     },
	//     Map2: map(<nil>),
	//     List: [
	//         interface {}(
	//             formatting.A{
	//                 VInt: int{10},
	//             }
	//         ),
	//         interface {}(
	//             int{10}
	//         ),
	//     ],
	//     List2: [
	//     ],
	//     F: func fmt.Printf(string, ...interface {}) (int, error),
	//     F2: (func(string, ...interface {}) (int, error))(<nil>),
	//     Ch: chan string{len=1,cap=20},
	// }

}
