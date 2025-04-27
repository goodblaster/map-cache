package main

import (
	"github.com/goodblaster/map-cache/internal/api"
	"github.com/goodblaster/map-cache/internal/config"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

func main() {
	config.Init()
	err := caches.AddCache(caches.DefaultName)
	if err != nil {
		panic(err)
	}

	//cache, err := caches.FetchCache(caches.DefaultName)
	//if err != nil {
	//	panic(err)
	//}
	//defer cache.Release(caches.DefaultName)
	//
	//// Initialize the cache with some data
	//m := map[string]any{
	//	"key1": "value1",
	//	"key2": "value2",
	//	"key3": map[string]any{
	//		"innerKey1": "innerValue1",
	//		"innerKey2": "innerValue2",
	//		"outerKey1": "outerValue1",
	//	},
	//	"key4": []any{"item1", "item2", "item3"},
	//	"key5": 12345,
	//}
	//ctx := context.Background()
	//_ = cache.Replace(ctx, m)
	//
	//container := gabs.Wrap(m)
	//
	//value, ok := container.Path("key5").Data().(int)
	//// value == 10.0, ok == true
	//fmt.Println(value, ok)
	//
	////value, ok = container.Search("outer", "inner", "value1").Data().(float64)
	////// value == 10.0, ok == true
	////
	////value, ok = container.Search("outer", "alsoInner", "array1", "1").Data().(float64)
	////// value == 40.0, ok == true
	////
	//gObj, err := container.JSONPointer("/key4/1")
	//if err != nil {
	//	panic(err)
	//}
	//v := gObj.Data()
	//fmt.Println(v)
	//
	//a, err := container.SetJSONPointer(40, "/key4/1")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(a)
	//
	//gObj, err = container.JSONPointer("/key4/1")
	//if err != nil {
	//	panic(err)
	//}
	//v = gObj.Data()
	//fmt.Println(v)
	//
	//gObj, err = container.JSONPointer("")
	//if err != nil {
	//	panic(err)
	//}
	//v = gObj.Data()
	//fmt.Println(v)
	//
	//children := container.Path("key3.inner*").Children()
	//fmt.Println(children)

	//container.S("foo").SetIndex("test2", 1)
	// value == 40.0, ok == true
	//
	//value, ok = container.Path("does.not.exist").Data().(float64)
	//// value == 0.0, ok == false
	//
	//exists := container.Exists("outer", "inner", "value1")
	//// exists == true
	//
	//exists = container.ExistsP("does.not.exist")
	//// exists == false

	e := echo.New()
	api.SetupRoutes(e)
	_ = e.Start(":8080")
}
