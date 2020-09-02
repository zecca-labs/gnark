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

package fields

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gurvy/bls377"
)

// Fp6Elmt element in a quadratic extension
type Fp6Elmt struct {
	B0, B1, B2 Fp2Elmt
}

func (e *Fp6Elmt) Assign(a *bls377.E6) {
	e.B0.Assign(&a.B0)
	e.B1.Assign(&a.B1)
	e.B2.Assign(&a.B2)
}

func (e *Fp6Elmt) MUSTBE_EQ(cs *frontend.CS, other Fp6Elmt) {
	e.B0.MUSTBE_EQ(cs, other.B0)
	e.B1.MUSTBE_EQ(cs, other.B1)
	e.B2.MUSTBE_EQ(cs, other.B2)
}

// Add creates a fp6elmt from fp elmts
func (e *Fp6Elmt) Add(cs *frontend.CS, e1, e2 *Fp6Elmt) *Fp6Elmt {

	e.B0.Add(cs, &e1.B0, &e2.B0)
	e.B1.Add(cs, &e1.B1, &e2.B1)
	e.B2.Add(cs, &e1.B2, &e2.B2)

	return e
}

// NewFp6Zero creates a new
func NewFp6Zero(cs *frontend.CS) Fp6Elmt {
	return Fp6Elmt{
		B0: Fp2Elmt{cs.ALLOCATE(0), cs.ALLOCATE(0)},
		B1: Fp2Elmt{cs.ALLOCATE(0), cs.ALLOCATE(0)},
		B2: Fp2Elmt{cs.ALLOCATE(0), cs.ALLOCATE(0)},
	}
}

// Sub creates a fp6elmt from fp elmts
func (e *Fp6Elmt) Sub(cs *frontend.CS, e1, e2 *Fp6Elmt) *Fp6Elmt {

	e.B0.Sub(cs, &e1.B0, &e2.B0)
	e.B1.Sub(cs, &e1.B1, &e2.B1)
	e.B2.Sub(cs, &e1.B2, &e2.B2)

	return e
}

// Neg negates an Fp6 elmt
func (e *Fp6Elmt) Neg(cs *frontend.CS, e1 *Fp6Elmt) *Fp6Elmt {
	e.B0.Neg(cs, &e1.B0)
	e.B1.Neg(cs, &e1.B1)
	e.B2.Neg(cs, &e1.B2)
	return e
}

// Mul creates a fp6elmt from fp elmts
// icube is the imaginary elmt to the cube
func (e *Fp6Elmt) Mul(cs *frontend.CS, e1, e2 *Fp6Elmt, ext Extension) *Fp6Elmt {

	// notations: (a+bv+cv2)*(d+ev+fe2)
	var ad, bf, ce Fp2Elmt
	ad.Mul(cs, &e1.B0, &e2.B0, ext)                       // 5C
	bf.Mul(cs, &e1.B1, &e2.B2, ext).MulByIm(cs, &bf, ext) // 6C
	ce.Mul(cs, &e1.B2, &e2.B1, ext).MulByIm(cs, &ce, ext) // 6C

	var cf, ae, bd Fp2Elmt
	cf.Mul(cs, &e1.B2, &e2.B2, ext).MulByIm(cs, &cf, ext) // 6C
	ae.Mul(cs, &e1.B0, &e2.B1, ext)                       // 5C
	bd.Mul(cs, &e1.B1, &e2.B0, ext)                       // 5C

	var af, be, cd Fp2Elmt
	af.Mul(cs, &e1.B0, &e2.B2, ext) // 5C
	be.Mul(cs, &e1.B1, &e2.B1, ext) // 5C
	cd.Mul(cs, &e1.B2, &e2.B0, ext) // 5C

	e.B0.Add(cs, &ad, &bf).Add(cs, &e.B0, &ce) // 4C
	e.B1.Add(cs, &cf, &ae).Add(cs, &e.B1, &bd) // 4C
	e.B2.Add(cs, &af, &be).Add(cs, &e.B2, &cd) // 4C

	return e
}

// MulByFp2 creates a fp6elmt from fp elmts
// icube is the imaginary elmt to the cube
func (e *Fp6Elmt) MulByFp2(cs *frontend.CS, e1 *Fp6Elmt, e2 *Fp2Elmt, ext Extension) *Fp6Elmt {
	res := Fp6Elmt{}

	res.B0.Mul(cs, &e1.B0, e2, ext)
	res.B1.Mul(cs, &e1.B1, e2, ext)
	res.B2.Mul(cs, &e1.B2, e2, ext)

	e.B0 = res.B0
	e.B1 = res.B1
	e.B2 = res.B2

	return e
}

// MulByNonResidue multiplies e by the imaginary elmt of Fp6 (noted a+bV+cV where V**3 in F^2)
func (e *Fp6Elmt) MulByNonResidue(cs *frontend.CS, e1 *Fp6Elmt, ext Extension) *Fp6Elmt {
	res := Fp6Elmt{}
	res.B0.Mul(cs, &e1.B2, &ext.vCube, ext)
	e.B1 = e1.B0
	e.B2 = e1.B1
	e.B0 = res.B0
	return e
}

// Inverse inverses an Fp2 elmt
func (e *Fp6Elmt) Inverse(cs *frontend.CS, e1 *Fp6Elmt, ext Extension) *Fp6Elmt {

	var t [7]Fp2Elmt
	var c [3]Fp2Elmt
	var buf Fp2Elmt

	t[0].Mul(cs, &e1.B0, &e1.B0, ext)
	t[1].Mul(cs, &e1.B1, &e1.B1, ext)
	t[2].Mul(cs, &e1.B2, &e1.B2, ext)
	t[3].Mul(cs, &e1.B0, &e1.B1, ext)
	t[4].Mul(cs, &e1.B0, &e1.B2, ext)
	t[5].Mul(cs, &e1.B1, &e1.B2, ext)

	c[0].MulByIm(cs, &t[5], ext)

	c[0].Neg(cs, &c[0]).Add(cs, &c[0], &t[0])

	c[1].MulByIm(cs, &t[2], ext)

	c[1].Sub(cs, &c[1], &t[3])
	c[2].Sub(cs, &t[1], &t[4])
	t[6].Mul(cs, &e1.B2, &c[1], ext)
	buf.Mul(cs, &e1.B1, &c[2], ext)
	t[6].Add(cs, &t[6], &buf)

	t[6].MulByIm(cs, &t[6], ext)

	buf.Mul(cs, &e1.B0, &c[0], ext)
	t[6].Add(cs, &t[6], &buf)

	t[6].Inverse(cs, &t[6], ext)
	e.B0.Mul(cs, &c[0], &t[6], ext)
	e.B1.Mul(cs, &c[1], &t[6], ext)
	e.B2.Mul(cs, &c[2], &t[6], ext)

	return e

}
