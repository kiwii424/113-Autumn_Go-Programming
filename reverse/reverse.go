package reverse

import (
    // "fmt"
    "unicode/utf8"
)

func reverse(b []byte) {
    i, j := 0, len(b)
    for i < j {
        _, size1 := utf8.DecodeRune(b[i:])
        _, size2 := utf8.DecodeLastRune(b[:j])

        if size1 > size2 {
            temp := make([]byte, 0)
            for len(temp) < size1 {
                _, size := utf8.DecodeLastRune(b[:j])
                temp = append([]byte(string(b[j-size:j])), temp...)
                j -= size
            }
            // fmt.Println("Temp content for size1 > size2:", temp)
            copy(b[j:j+size1], b[i:i+size1])
            copy(b[i:i+size1], temp)
            i += size1
        } else if size2 > size1 {
            temp := make([]byte, 0)
            for len(temp) < size2 {
                _, size := utf8.DecodeRune(b[i:])
                temp = append([]byte(string(b[i:i+size])), temp...)
                i += size
            }
            // fmt.Println("Temp content for size2 > size1:", temp)
            copy(b[i-len(temp):i], b[j-size2:j])
            copy(b[j-size2:j], temp)
            j -= size2
        } else {
            for k := 0; k < size1; k++ {
                b[i+k], b[j-size1+k] = b[j-size1+k], b[i+k]
            }
            i += size1
            j -= size2
        }
    }
}
