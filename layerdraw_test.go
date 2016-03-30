package main

import "testing"

func TestLayerContainerAdded(t *testing.T) {
	layer := NewLayer()
	container := NewContainer(&Dimensions{0, 0, 1, 1})
	layer.Add(container)
	if len(layer.Containers) != 1 {
		t.Fatal("len(layer.Containers) != 1")
	}
}

func TestAddWord(t *testing.T) {
	container := NewContainer(&Dimensions{0, 0, 1, 1})
	container.Add(NewWordDef("Test", 10))
	container.Add(NewWordDef("Test2", 10))
	if len(container.ContainerElements) != 2 {
		t.Fatal(len(container.ContainerElements) != 2)
	}
}

func TestWordPrintedPadding(t *testing.T) {
	layer := NewLayer()
	testdata := []struct {
		Word   string
		Length int
	}{
		{"Test", 10},
		{"More", 10},
	}
	container := NewContainer(&Dimensions{0, 0, 20, 1})
	expected := make([]byte, 0)
	for _, v := range testdata {
		container.Add(NewWordDef(v.Word, v.Length))
		for i := 0; i < v.Length; i++ {
			if len(v.Word) > i {
				expected = append(expected, byte(v.Word[i]))
			} else {
				expected = append(expected, byte(' '))
			}
		}
	}

	layer.Add(container)
	layer.Draw()
	buf := layer.GetBuff()
	t.Log(buf)
	if len(buf) != len(expected) {
		t.Fatal("len(buf) != len(expected):", len(buf), len(expected))
	}
	for k, v := range buf {
		if byte(v.Char) != expected[k] {
			t.Fatal("if byte(v.Char) != expected[k]", string(v.Char), k)
		}
	}

}
