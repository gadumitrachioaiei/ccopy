package ccopy

import "fmt"

func Example() {
	// we want a copy of below object, but we want to change the name
	type T struct {
		A    int
		Name string `ccopy:"anonymiseName"`
	}
	obj := T{A: 2, Name: "Secret name"}

	// a config function that receives some instance of data tagged using our tag, and returns changed data, of same type
	anonymiseName := func(name string) string {
		return "john doe"
	}

	c := Config{"anonymiseName": anonymiseName}
	objCopy, err := c.Copy(obj)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", objCopy)
	// Output:
	// ccopy.T{A:2, Name:"john doe"}
}
