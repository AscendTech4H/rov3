package main

import "testing"

func TestVec3Unit(t *testing.T) {
	if (Vec3{29183284.2374234, 2738223.232732, -2372}.Unit().Magnitude() - 1) > 0.0001 {
		t.Fatalf("Magnitude of unit vector is not 1\n")
	}
}

func TestXYZ(t *testing.T) {
	if (Vec3{1, 0, 0}.X()) != 1 {
		t.Fatalf("X does not work")
	}
	if (Vec3{0, 1, 0}.Y()) != 1 {
		t.Fatalf("Y does not work")
	}
	if (Vec3{0, 0, 1}.Z()) != 1 {
		t.Fatalf("Z does not work")
	}
}

func TestIJKUnit(t *testing.T) {
	if I.Unit() != I {
		t.Fatal("Unit vector of i is not i")
	}
	if J.Unit() != J {
		t.Fatal("Unit vector of i is not i")
	}
	if K.Unit() != K {
		t.Fatal("Unit vector of i is not i")
	}
}

func TestAdd(t *testing.T) {
	v := I.Add(J).Add(K)
	if v.X() != v.Y() || v.Y() != v.Z() {
		t.Fatalf("Bad addition: expected <1, 1, 1>, but got <%f, %f, %f>", v.X(), v.Y(), v.Z())
	}
}
