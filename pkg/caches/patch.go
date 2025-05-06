package caches

//import (

//)
//
//type Patch struct {
//	Operation  string      `json:"operation"`
//	Key        string      `json:"key"`
//	Value      any         `json:"value"`
//	Conditions []Condition `json:"conditions,omitempty"`
//}
//
//type Condition struct {
//	If   Comparison `json:"if"`
//	Then Patch      `json:"then"`
//	Else Patch      `json:"else"`
//}
//
//type Comparison struct {
//	Key      string `json:"key"`
//	Operator string `json:"operator"`
//	Value    any    `json:"value"`
//}
//
//func (cache Cache) Patch(ctx context.Context, patches ...Patch) error {
//	// I would prefer this be atomic. But, for now ...
//	//
//
//	for _, patch := range patches {
//		if err := cache.ApplyPatch(patch); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (cache Cache) ApplyPatch(patch Patch) error {
//	switch patch.Operation {
//	case "set":
//		caval, err := cache.Navigate(patch.Key)
//		if err != nil {
//			return err
//		}
//		caval.Set(patch.Value)
//
//	case "del":
//		caval, err := cache.Navigate(patch.Key)
//		if err != nil {
//			break // already deleted
//		}
//		_ = caval.Delete()
//
//	case "inc":
//		caval, err := cache.Navigate(patch.Key)
//		if err != nil {
//			return err
//		}
//
//		// patch.Value must be a number. float64, specifically.
//		f64, ok := patch.Value.(float64)
//		if !ok {
//			return fmt.Errorf("value must be a number") // todo error
//		}
//
//		if err := caval.Increment(f64); err != nil {
//			return err
//		}
//
//	case "dec":
//		caval, err := cache.Navigate(patch.Key)
//		if err != nil {
//			return err
//		}
//
//		// patch.Value must be a number. float64, specifically.
//		f64, ok := patch.Value.(float64)
//		if !ok {
//			return fmt.Errorf("value must be a number") // todo error
//		}
//
//		if err := caval.Increment(-f64); err != nil {
//			return err
//		}
//
//	default:
//		return errors.New("not implemented")
//	}
//
//	//// Handle conditions
//	//for _, cond := range patch.Conditions {
//	//	if EvaluateCondition(data, cond.If) {
//	//		ApplyPatch(data, cond.Then)
//	//	} else {
//	//		ApplyPatch(data, cond.Else)
//	//	}
//	//}
//
//	return nil
//}
//
//// Tons to do here.
//func (cache Cache) EvaluateCondition(comp Comparison) (bool, error) {
//	caval, err := cache.Navigate(comp.Key)
//	if err != nil {
//		return false, err
//	}
//
//	current, ok := caval.Get()
//	if !ok {
//		return false, ErrKeyNotFound
//	}
//
//	switch comp.Operator {
//	case ">":
//		return compareFloat(current, comp.Value) > 0, nil
//	case "<":
//		return compareFloat(current, comp.Value) < 0, nil
//	case "==":
//		return current == comp.Value, nil
//	case "!=":
//		return current != comp.Value, nil
//	}
//	return false, nil
//}
//
//func compareFloat(a, b any) float64 {
//	fa, _ := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
//	fb, _ := strconv.ParseFloat(fmt.Sprintf("%v", b), 64)
//	return fa - fb
//}
