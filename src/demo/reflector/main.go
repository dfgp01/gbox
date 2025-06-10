package main

import "gbox/demo/reflector/internal"

func main() {
	//internal.RunNative()
	// internal.RunFrame(internal.WrongVar)
	//internal.RunFrame(internal.HackVar)
	// internal.RunFrame(internal.Str)
	// internal.RunFrame(internal.SlicePStruct)
	// internal.RunFrame(internal.SimpleMap)
	// internal.RunFrame(internal.NormalMap)
	// internal.RunFrame(internal.HardMap)

	internal.ReflectTest3()
}
