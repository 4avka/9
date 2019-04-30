package feck

import (
	"crypto/rand"
	"testing"
)

const (
	requiredshards = 3
	totalshards    = 9
)

func TestSplit(t *testing.T) {
	testData := make([]byte, 256)
	_, _ = rand.Read(testData)
	t.Log(len(testData), testData)
	rsc := New(requiredshards, totalshards)
	encoded := rsc.Encode(testData)
	for i, x := range encoded {
		t.Log(i, len(x), x)
	}
	muddlemap := make(map[int][]byte)
	for i, x := range encoded {
		muddlemap[i] = x
	}
	var muddled [][]byte
	for _, x := range muddlemap {
		muddled = append(muddled, x)
	}
	for i, x := range muddled {
		t.Log(i, len(x), x)
		if i%2 == 0 || i%3 == 0 {
			muddled[i][int(muddled[i][4])/len(muddled[i])] = ^muddled[i][int(muddled[i][4])/len(muddled[i])]
			t.Log("MANGLING", i, len(x), x)
		}
	}
	unmuddled := rsc.Decode(muddled)
	t.Log(unmuddled)
	// padded := PadData(testData, 3, 9)
	// t.Log(len(padded), padded)
	// splitted := Split(padded, 3, 9)

	// err := rsc.Reconst(splitted, []int{0, 1, 2}, []int{3, 4, 5, 6, 7, 8})
	// if err != nil {
	// 	panic(err)
	// }
	// for i, x := range splitted {
	// 	t.Log(i, len(x), x)
	// }
	// for i := 0; i < 6; i++ {
	// 	ZeroBytes(splitted[i])
	// }
	// err = rsc.Reconst(splitted, []int{6, 7, 8}, []int{0, 1, 2})
	// if err != nil {
	// 	panic(err)
	// }
	// t.Log("only last 3 generated polynomials were intact")
	// for i, x := range splitted {
	// 	t.Log(i, len(x), x)
	// }
	// err = rsc.Reconst(splitted, []int{0, 1, 2}, []int{3, 4, 5})
	// if err != nil {
	// 	panic(err)
	// }
	// for i, x := range splitted {
	// 	t.Log(i, len(x), x)
	// }
	// for i, x := range splitted {
	// 	splitted[i] = append([]byte{byte(i)}, x...)
	// 	splitted[i] = AppendChecksum(splitted[i])
	// }
	// for i := 0; i < 6; i++ {
	// 	splitted[i][len(splitted[i])-5] = 0
	// }
	// for i, x := range splitted {
	// 	t.Log(i, len(x), x)
	// }
	// var have, missing []int
	// var verified bool
	// for i, x := range splitted {
	// 	splitted[i], verified = VerifyChecksum(x)
	// 	if verified {
	// 		have = append(have, i)
	// 	} else {
	// 		missing = append(missing, i)
	// 	}
	// }
	// t.Log(have, missing)
	// t.Log(">>> first six below should be zeroed due to prior mangling")
	// for i := 6; i < 9; i++ {
	// 	splitted[i] = splitted[i][1:]
	// }
	// for i, x := range splitted {
	// 	t.Log(i, len(x), x)
	// }
	// err = rsc.Reconst(splitted, []int{6, 7, 8}, []int{0, 1, 2})
	// if err != nil {
	// 	panic(err)
	// }
	// for i, x := range splitted {
	// 	t.Log(i, len(x), x)
	// }
	// result := append(splitted[0], append(splitted[1], splitted[2]...)...)
	// t.Log(result)
	// unpadded := UnpadData(result)
	// t.Log(unpadded)
	// t.Log(testData)
	// t.Log(string(unpadded) == string(testData))
}
