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

package frontend

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/consensys/gnark/backend"
)

// ADD Adds 2+ inputs and returns resulting Constraint
func (cs *CS) ADD(i1, i2 interface{}, in ...interface{}) Variable {

	// can add constraint and constants
	add := func(_i1, _i2 interface{}) Variable {
		switch c1 := _i1.(type) {
		case Variable:
			switch c2 := _i2.(type) {
			case Variable:
				return cs.add(c1, c2)
			default:
				return cs.addConstant(c1, backend.FromInterface(c2))
			}
		default:
			switch c2 := _i2.(type) {
			case Variable:
				return cs.addConstant(c2, backend.FromInterface(c1))
			default:
				panic("invalid type")
			}
		}
	}

	res := add(i1, i2)

	for i := 0; i < len(in); i++ {
		res = add(res, in[i])
	}

	return res
}

// SUB Adds two constraints
func (cs *CS) SUB(i1, i2 interface{}) Variable {
	switch c1 := i1.(type) {
	case Variable:
		switch c2 := i2.(type) {
		case Variable:
			return cs.sub(c1, c2)
		case big.Int:
			return cs.subConstant(c1, c2)
		}
	default:
		_c1 := backend.FromInterface(c1)
		switch c2 := i2.(type) {
		case Variable:
			return cs.subConstraint(_c1, c2)
		}
	}
	panic("invalid type")
}

// MUL Multiplies 2+ constraints together
func (cs *CS) MUL(i1, i2 interface{}, in ...interface{}) Variable {

	// multiplies 2 terms (constraints, Elements, uint64, int, String)
	mul := func(_i1, _i2 interface{}) Variable {
		switch c1 := _i1.(type) {
		case LinearCombination:
			switch c2 := _i2.(type) {
			case LinearCombination:
				return cs.mullc(c1, c2)
			default:
				return cs.mullcinterface(c1, c2)
			}
		case Variable:
			switch c2 := _i2.(type) {
			case Variable:
				return cs.mul(c1, c2)
			case LinearCombination:
				return cs.mullcinterface(c2, c1)
			default:
				return cs.mulConstant(c1, backend.FromInterface(c2))
			}
		default: // i1 is not a Constraint type, so c2 must be
			switch c2 := _i2.(type) {
			case Variable:
				return cs.mulConstant(c2, backend.FromInterface(c1))
			default:
				fmt.Println(reflect.TypeOf(_i2))
				panic("invalid type")
			}

		}
	}

	res := mul(i1, i2)

	for i := 0; i < len(in); i++ {
		res = mul(res, in[i])
	}

	return res

}

// DIV divides two constraints (i1/i2)
func (cs *CS) DIV(i1, i2 interface{}) Variable {

	div := func(_i1, _i2 interface{}) Variable {
		switch c1 := _i1.(type) {
		case LinearCombination:
			switch c2 := _i2.(type) {
			case LinearCombination:
				return cs.divlc(c1, c2)
			default:
				panic("invalid type; only support linear expression DIV linear expression")
			}
		case Variable:
			switch c2 := _i2.(type) {
			case Variable:
				return cs.div(c1, c2)
			default:
				tmp := backend.FromInterface(c2)
				return cs.divConstantRight(c1, tmp)
			}
		default: // i1 is not a Constraint type, so c2 must be
			switch c2 := _i2.(type) {
			case Variable:
				tmp := backend.FromInterface(c1)
				return cs.divConstantLeft(tmp, c2)
			default:
				panic("invalid type")
			}

		}
	}

	res := div(i1, i2)

	return res

}

// MUSTBE_EQ equalizes two constraints
func (cs *CS) MUSTBE_EQ(i1, i2 interface{}) {

	switch c1 := i1.(type) {
	case Variable:
		switch c2 := i2.(type) {
		case Variable:
			cs.equal(c1, c2)
			return
		default:
			b2 := backend.FromInterface(i2)
			cs.equalConstant(c1, b2)
			return
		}
	default:
		b1 := backend.FromInterface(i1)
		switch c2 := i2.(type) {
		case Variable:
			cs.equalConstant(c2, b1)
			return
		}
	}
	panic("invalid type: MUSTBE_EQ takes Variables and big.Int only")
}

// INV inverse a Constraint
func (cs *CS) INV(c1 Variable) Variable {
	return cs.inv(c1, *bOne)
}

// XOR compute the xor between two constraints
func (cs *CS) XOR(c1, c2 Variable) Variable {
	// ensure c1 and c2 are already boolean constrained
	cs.MUSTBE_BOOLEAN(c1)
	cs.MUSTBE_BOOLEAN(c2)

	expression := xorExpression{
		a: c1.id(),
		b: c2.id(),
	}

	return cs.addConstraint(&expression)
}

// MUSTBE_BOOLEAN boolean constrains a variable
func (cs *CS) MUSTBE_BOOLEAN(c Variable) {
	// check if the variable is already boolean constrained
	for i := 0; i < len(cs.noExpressions); i++ {
		if bExpression, ok := cs.noExpressions[i].(*booleanExpression); ok {
			if bExpression.b == c.id() {
				// this variable is already boolean constrained
				return
			}
		}
	}
	// check if the variable is the result of a XOR (a xor b == c --> c is automatically boolean constrained)
	for cID, val := range cs.constraints {
		if cID == 0 {
			continue // skipping first entry, reserved
		}
		if cID == c.id() {
			if _, ok := val.exp.(*xorExpression); ok {
				// constraint is the result of a xor expression and is already boolean constrained as such
				return
			}
		}
	}
	cs.noExpressions = append(cs.noExpressions, &booleanExpression{b: c.id()})
}

// TO_BINARY unpacks a variable in binary, n is the number of bits of the variable
// The result in in little endian (first bit= lsb)
func (cs *CS) TO_BINARY(c Variable, nbBits int) []Variable {

	// create the expression ensuring the bit decomposition matches c
	expression := &unpackExpression{
		res: c.id(),
	}
	cs.moExpressions = append(cs.moExpressions, expression)

	// create our bits constraints
	bits := make([]Variable, nbBits)
	for i := 0; i < nbBits; i++ {
		bits[i] = cs.addConstraint(nil)
		cs.MUSTBE_BOOLEAN(bits[i]) // (MUSTBE_BOOLEAN check for duplicate constraints)
		expression.bits = append(expression.bits, bits[i].id())
	}

	return bits
}

// FROM_BINARY packs b, seen as a fr.Element in little endian
func (cs *CS) FROM_BINARY(b ...Variable) Variable {

	expression := packExpression{}

	for _, c := range b {
		cs.MUSTBE_BOOLEAN(c) // ensure input is boolean constrained
		expression.bits = append(expression.bits, c.id())
	}

	return cs.addConstraint(&expression)
}

// MUSTBE_LESS_OR_EQ constrains c to be less or equal than e (taken as lifted Integer values from Fr)
// from https://github.com/zcash/zips/blob/master/protocol/protocol.pdf
func (cs *CS) MUSTBE_LESS_OR_EQ(c Variable, bound interface{}, nbBits int) {

	switch _bound := bound.(type) {
	case Variable:
		cs.mustBeLessOrEq(c, _bound, nbBits)
	default:
		b := backend.FromInterface(bound)
		cs.mustBeLessOrEqConstant(c, b, nbBits)
	}
}

// SELECT if b is true, yields c1 else yields c2
func (cs *CS) SELECT(b Variable, i1, i2 interface{}) Variable {

	// ensure b is boolean constrained
	cs.MUSTBE_BOOLEAN(b)

	switch c1 := i1.(type) {
	case Variable:
		switch c2 := i2.(type) {
		case Variable:
			expression := selectExpression{
				b: b.id(),
				x: c1.id(),
				y: c2.id(),
			}
			return cs.addConstraint(&expression)
		default:
			panic("invalid type")
		}
	default:
		c1Bigint := backend.FromInterface(i1)
		c2Bigint := backend.FromInterface(i2)
		c1Bigint.Sub(&c1Bigint, &c2Bigint)
		expression := linearExpression{
			cs.term(b.id(), c1Bigint),
			cs.term(oneWireID, *bOne),
		}
		return cs.addConstraint(&expression)
	}
}

// SELECT_LUT select lookuptable[c1*2+c0] where c0 and c1 are boolean constrained
// cf https://z.cash/technology/jubjub/
func (cs *CS) SELECT_LUT(c1, c0 Variable, lookuptable [4]big.Int) Variable {

	// ensure c0 and c1 are boolean constrained
	cs.MUSTBE_BOOLEAN(c0)
	cs.MUSTBE_BOOLEAN(c1)

	expression := lutExpression{
		b0:          c0.id(),
		b1:          c1.id(),
		lookuptable: lookuptable,
	}

	return cs.addConstraint(&expression)

}

// SECRET_INPUT creates a Constraint containing an input
func (cs *CS) SECRET_INPUT(name string) Variable {
	if name == "" {
		panic("can't initialize nameless input")
	}

	v := cs.addConstraint(nil)
	// ensure name is not duplicate
	if _, ok := cs.wireNames[name]; ok {
		panic("duplicate input name")
	} else {
		cs.wireNames[name] = struct{}{}
		cs.secretWireNames[v.id()] = name
	}

	return v

}

// PUBLIC_INPUT creates a Constraint containing an input
func (cs *CS) PUBLIC_INPUT(name string) Variable {
	if name == "" {
		panic("can't initialize nameless input")
	}

	v := cs.addConstraint(nil)
	// ensure name is not duplicate
	if _, ok := cs.wireNames[name]; ok {
		panic("duplicate input name")
	} else {
		cs.wireNames[name] = struct{}{}
		cs.publicWireNames[v.id()] = name
	}

	return v
}

// ALLOCATE will return an allocated cs.Constraint from input {Constraint, element, uint64, int, ...}
func (cs *CS) ALLOCATE(input interface{}) Variable {
	switch x := input.(type) {
	case Variable:
		return x
	default:
		return cs.constVar(x)
	}
}

// Tag tags variable v with tag
// useful for debug purposes, and retrieve intermediate values once through r1cs.Inspect() method
func (cs *CS) Tag(v Variable, tag string) {
	for _, v := range cs.wireTags {
		for _, t := range v {
			if tag == t {
				panic("duplicate tag " + tag)
			}
		}
	}
	cID := v.id()
	tags := cs.wireTags[cID]
	tags = append(tags, tag)
	cs.wireTags[cID] = tags
}
