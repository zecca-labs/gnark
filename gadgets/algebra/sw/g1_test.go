/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sw

import (
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gurvy"
	"github.com/consensys/gurvy/bls377/fp"
	"github.com/consensys/gurvy/bls377/fr"

	"github.com/consensys/gurvy/bls377"
)

//--------------------------------------------------------------------
// utils

func randomPointG1() bls377.G1Jac {

	curve := bls377.BLS377()

	var p1 bls377.G1Jac

	p1.X.SetString("68333130937826953018162399284085925021577172705782285525244777453303237942212457240213897533859360921141590695983")
	p1.Y.SetString("243386584320553125968203959498080829207604143167922579970841210259134422887279629198736754149500839244552761526603")
	p1.Z.SetString("1")

	var r1 fr.Element
	r1.SetRandom()

	p1.ScalarMul(curve, &p1, r1)

	return p1
}

func newPointCircuitG1(circuit *frontend.CS, s string) *G1Jac {
	return NewPointG1(circuit,
		circuit.SECRET_INPUT(s+"0"),
		circuit.SECRET_INPUT(s+"1"),
		circuit.SECRET_INPUT(s+"2"),
	)
}

func newPointCircuitG1Aff(circuit *frontend.CS, s string) *G1Aff {
	return NewPointG1Aff(circuit,
		circuit.SECRET_INPUT(s+"0"),
		circuit.SECRET_INPUT(s+"1"),
	)
}

func tagPointG1(cs *frontend.CS, g *G1Jac, s string) {
	cs.Tag(g.X, s+"0")
	cs.Tag(g.Y, s+"1")
	cs.Tag(g.Z, s+"2")
}

func tagPointG1Aff(cs *frontend.CS, g *G1Aff, s string) {
	cs.Tag(g.X, s+"0")
	cs.Tag(g.Y, s+"1")
}

func assignPointG1(inputs map[string]interface{}, g bls377.G1Jac, s string) {
	inputs[s+"0"] = g.X.String()
	inputs[s+"1"] = g.Y.String()
	inputs[s+"2"] = g.Z.String()

}

func getExpectedValuesG1(m map[string]*fp.Element, s string, g bls377.G1Jac) {
	m[s+"0"] = &g.X
	m[s+"1"] = &g.Y
	m[s+"2"] = &g.Z
}

//--------------------------------------------------------------------
// test

func TestAddAssignG1(t *testing.T) {

	curve := bls377.BLS377()

	// sample 2 random points
	g1 := randomPointG1()
	g2 := randomPointG1()

	// create the circuit
	circuit := frontend.NewConstraintSystem()

	gc1 := newPointCircuitG1(&circuit, "a")
	gc2 := newPointCircuitG1(&circuit, "b")
	gc1.AddAssign(&circuit, gc2)
	tagPointG1(&circuit, gc1, "c")

	// assign the inputs
	inputs := make(map[string]interface{})
	assignPointG1(inputs, g1, "a")
	assignPointG1(inputs, g2, "b")

	// compute the result
	g1.Add(curve, &g2)

	// assign the exepected values
	expectedValues := make(map[string]*fp.Element)
	getExpectedValuesG1(expectedValues, "c", g1)

	// check expected result
	r1cs := circuit.ToR1CS().ToR1CS(gurvy.BW761)

	res, err := r1cs.Inspect(inputs, false)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range res {
		_v := fp.FromInterface(v)
		if !expectedValues[k].Equal(&_v) {
			t.Fatal("error add g1")
		}
	}
}

func TestAddAssignAffG1(t *testing.T) {

	curve := bls377.BLS377()

	// sample 2 random points
	g1 := randomPointG1()
	g2 := randomPointG1()
	var _g1, _g2 bls377.G1Affine
	g1.ToAffineFromJac(&_g1)
	g2.ToAffineFromJac(&_g2)

	// create the circuit
	circuit := frontend.NewConstraintSystem()

	gc1 := newPointCircuitG1Aff(&circuit, "a")
	gc2 := newPointCircuitG1Aff(&circuit, "b")
	gc1.AddAssign(&circuit, gc2)
	tagPointG1Aff(&circuit, gc1, "c")

	// assign the inputs
	var one fp.Element
	one.SetUint64(1)
	inputs := make(map[string]interface{})
	inputs["a0"] = _g1.X.String()
	inputs["a1"] = _g1.Y.String()

	inputs["b0"] = _g2.X.String()
	inputs["b1"] = _g2.Y.String()

	// compute the result
	var _gres bls377.G1Affine
	g1.Add(curve, &g2)
	g1.ToAffineFromJac(&_gres)

	// assign the exepected values
	expectedValues := make(map[string]*fp.Element)
	expectedValues["c0"] = &_gres.X
	expectedValues["c1"] = &_gres.Y

	// check expected result
	r1cs := circuit.ToR1CS().ToR1CS(gurvy.BW761)

	res, err := r1cs.Inspect(inputs, false)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range res {
		_v := fp.FromInterface(v)
		if !expectedValues[k].Equal(&_v) {
			t.Fatal("error add g1 (affine)")
		}
	}
}

func TestDoubleG1(t *testing.T) {

	// sample 2 random points
	g1 := randomPointG1()

	// create the circuit
	circuit := frontend.NewConstraintSystem()

	gc1 := newPointCircuitG1(&circuit, "a")
	gc1.DoubleAssign(&circuit)
	tagPointG1(&circuit, gc1, "c")

	// assign the inputs
	inputs := make(map[string]interface{})
	assignPointG1(inputs, g1, "a")

	// compute the result
	g1.Double()

	// assign the exepected values
	expectedValues := make(map[string]*fp.Element)
	getExpectedValuesG1(expectedValues, "c", g1)

	// check expected result
	r1cs := circuit.ToR1CS().ToR1CS(gurvy.BW761)
	res, err := r1cs.Inspect(inputs, false)

	if err != nil {
		t.Fatal(err)
	}
	for k, v := range res {
		_v := fp.FromInterface(v)
		if !expectedValues[k].Equal(&_v) {
			t.Fatal("error double g1")
		}
	}
}

func TestDoubleAffG1(t *testing.T) {

	// sample a random points
	g1 := randomPointG1()
	var _g1 bls377.G1Affine
	g1.ToAffineFromJac(&_g1)

	// create the circuit
	circuit := frontend.NewConstraintSystem()

	gc1 := newPointCircuitG1Aff(&circuit, "a")
	gc1.Double(&circuit, gc1)
	tagPointG1Aff(&circuit, gc1, "c")

	// assign the inputs
	inputs := make(map[string]interface{})
	inputs["a0"] = _g1.X.String()
	inputs["a1"] = _g1.Y.String()

	// compute the reference result
	var _gres bls377.G1Affine
	g1.Double()
	g1.ToAffineFromJac(&_gres)

	// assign the exepected values
	expectedValues := make(map[string]*fp.Element)
	expectedValues["c0"] = &_gres.X
	expectedValues["c1"] = &_gres.Y

	// check expected result
	r1cs := circuit.ToR1CS().ToR1CS(gurvy.BW761)

	res, err := r1cs.Inspect(inputs, false)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range res {
		_v := fp.FromInterface(v)
		if !expectedValues[k].Equal(&_v) {
			t.Fatal("error double g1 (affine)")
		}
	}
}

func TestNegG1(t *testing.T) {

	// sample 2 random points
	g1 := randomPointG1()

	// create the circuit
	circuit := frontend.NewConstraintSystem()

	gc1 := newPointCircuitG1(&circuit, "a")
	gc1.Neg(&circuit, gc1)
	tagPointG1(&circuit, gc1, "c")

	// assign the inputs
	inputs := make(map[string]interface{})
	assignPointG1(inputs, g1, "a")

	// compute the result
	g1.Neg(&g1)

	// assign the exepected values
	expectedValues := make(map[string]*fp.Element)
	getExpectedValuesG1(expectedValues, "c", g1)

	// check expected result
	r1cs := circuit.ToR1CS().ToR1CS(gurvy.BW761)
	res, err := r1cs.Inspect(inputs, false)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range res {
		_v := fp.FromInterface(v)
		if !expectedValues[k].Equal(&_v) {
			t.Fatal("error neg g1")
		}
	}

}

func TestScalarMulG1(t *testing.T) {

	curve := bls377.BLS377()

	// sample a random points
	g1 := randomPointG1()
	var g1Aff bls377.G1Affine
	g1.ToAffineFromJac(&g1Aff)

	// random scalar
	var r fr.Element
	r.SetRandom()

	// create the circuit
	circuit := frontend.NewConstraintSystem()
	gc1 := newPointCircuitG1Aff(&circuit, "gc1")
	gc1.ScalarMul(&circuit, gc1, r.String(), 256)
	tagPointG1Aff(&circuit, gc1, "res")

	// assign the inputs
	inputs := make(map[string]interface{})
	inputs["gc10"] = g1Aff.X.String()
	inputs["gc11"] = g1Aff.Y.String()

	// compute the result
	r.FromMont()
	g1.ScalarMul(curve, &g1, r)
	g1.ToAffineFromJac(&g1Aff)

	// assign the exepected values
	expectedValues := make(map[string]*fp.Element)
	expectedValues["res0"] = &g1Aff.X
	expectedValues["res1"] = &g1Aff.Y

	// check expected result
	r1cs := circuit.ToR1CS().ToR1CS(gurvy.BW761)

	res, err := r1cs.Inspect(inputs, false)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range res {
		_v := fp.FromInterface(v)
		if !expectedValues[k].Equal(&_v) {
			t.Fatal("error scalar mul g1")
		}
	}
}
