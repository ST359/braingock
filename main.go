package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
)

func calculateElegantDivisor(val int) int {

	divisor := int(math.Ceil(math.Sqrt(float64(val))))
	for divisor > 1 {
		if val%divisor == 0 {
			break
		}
		divisor -= 1
	}
	return divisor
}

type Converter struct {
	currCell        int
	currLoopCounter int
	mem             map[int]int
	charset         map[int]struct{}
	outBuf          bytes.Buffer
}

func (c *Converter) output(w io.Writer) {
	w.Write(c.outBuf.Bytes())
}
func (c *Converter) incrementCurrentCellBy(diff int) {
	//TODO: need to calculate difference between current loop counter and divisor and change loop counter by this difference
	divisor := calculateElegantDivisor(diff)
	quotient := int(diff / divisor)
	//point to counter cell
	for range c.currCell {
		c.outBuf.WriteByte('<')
	}
	for range divisor {
		c.outBuf.WriteByte('+')
	}
	c.outBuf.WriteByte('[')
	//walk from counter cell back to cell which we need to increment
	for range c.currCell {
		c.outBuf.WriteByte('>')
	}
	//"multiply"
	for range quotient {
		c.outBuf.WriteByte('+')
	}
	//point to counter cell to decrement it
	for range c.currCell {
		c.outBuf.WriteByte('<')
	}
	c.outBuf.Write([]byte{'-', ']'})
	//return to currCell
	for range c.currCell {
		c.outBuf.WriteByte('>')
	}
}
func average(i []int) int {
	sum := 0
	for _, v := range i {
		sum += v
	}
	return int(sum / len(i))
}
func (c *Converter) incrementCellsToValue(v ...int) {
	numOfCells := len(v)
	avgVal := average(v)
	//calculate the most elegant divisor for avgVal
	//it is needed to make number of "+" outside loop
	//roghly equal to + inside loop
	divisor := calculateElegantDivisor(avgVal)
	if divisor == 1 {
		avgVal++
		divisor = calculateElegantDivisor(avgVal)
	}
	//prepare loop repeats
	for range divisor {
		c.outBuf.WriteByte('+')
	}
	//dont forget to increment and decrement current cell
	c.outBuf.Write([]byte{'[', '>'})
	c.currCell++
	increments := avgVal / divisor
	for range numOfCells {
		for range increments {
			c.outBuf.WriteByte('+')
		}
		c.outBuf.WriteByte('>')
		c.currCell++
	}
	for c.currCell > 0 {
		c.outBuf.WriteByte('<')
		c.currCell -= 1
	}
	c.outBuf.Write([]byte{'-', ']'})
	for i, val := range v {
		if val > avgVal {
			c.currCell++
			c.outBuf.WriteByte('>')
			if val-avgVal >= 15 {
				c.incrementCurrentCellBy(val - avgVal)
			} else {
				for range val - avgVal {
					c.outBuf.WriteByte('+')
				}
			}
		} else if val < avgVal {
			c.currCell++
			c.outBuf.WriteByte('>')
			for range avgVal - val {
				c.outBuf.WriteByte('-')
			}
		}
		c.mem[val] = i + 1 //0-th cell used for init loop
	}
}

func (c *Converter) preparePrintingPart(input string) {
	for _, v := range input {
		target := c.mem[int(v)]
		if c.currCell > target {
			diff := c.currCell - target
			for range diff {
				c.currCell--
				c.outBuf.WriteByte('<')
			}
		} else if c.currCell < target {
			diff := target - c.currCell
			for range diff {
				c.currCell++
				c.outBuf.WriteByte('>')
			}
		}
		c.outBuf.WriteByte('.')
	}
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func gcdMap(m map[int]struct{}) int {
	result := 0
	for k := range m {
		if result == 0 {
			result = k
		}
		result = gcd(result, k)
	}
	return result
}

func main() {
	var input string
	fmt.Println("Type something to convert into brainfuck code")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input = scanner.Text()
	}
	fmt.Print(input)
	conv := new(Converter)
	conv.charset = make(map[int]struct{})
	conv.mem = make(map[int]int)
	for _, v := range input {
		fmt.Print(int(v))
		conv.charset[int(v)] = struct{}{}
	}
	charslc := make([]int, 0)
	for v := range conv.charset {
		charslc = append(charslc, v)
	}
	fmt.Println(charslc)
	conv.incrementCellsToValue(charslc...)
	conv.preparePrintingPart(input)
	conv.output(os.Stdout)
}
