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
	for divisor >= 1 {
		if val%divisor == 0 {
			return divisor
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
	_, err := w.Write(c.outBuf.Bytes())
	if err != nil {
		fmt.Println(err.Error())
	}
}
func (c *Converter) incrementCurrentCellBy(diff int) {
	//TODO: here or somewhere else there is an error: asdfgh outputs odhfga
	//maybe try to ease the "optimizations" a bit
	absDiff := int(math.Abs(float64(diff)))
	divisor := calculateElegantDivisor(absDiff)
	diffWasIncremented := false
	if divisor == 1 {
		absDiff++
		divisor = calculateElegantDivisor(absDiff)
		diffWasIncremented = true
	}
	quotient := int(absDiff / divisor)
	//point to counter cell, however, if last action was to walk from pointer to cell, just remove x of >'s
	recentChars := c.outBuf.Bytes()[c.outBuf.Len()-c.currCell:]
	if string(recentChars) == string(bytes.Repeat([]byte{'>'}, c.currCell)) {
		c.outBuf.Truncate(c.outBuf.Len() - c.currCell)
	} else {
		for range c.currCell {
			c.outBuf.WriteByte('<')
		}
	}
	//we need to make loop counter equal to *divisor*
	for c.currLoopCounter != divisor {
		if c.currLoopCounter > divisor {
			c.outBuf.WriteByte('-')
			c.currLoopCounter--
		} else if c.currLoopCounter < divisor {
			c.outBuf.WriteByte('+')
			c.currLoopCounter++
		}
	}
	c.outBuf.WriteByte('[')
	//walk from counter cell back to cell which we need to increment
	for range c.currCell {
		c.outBuf.WriteByte('>')
	}
	//"multiply"
	for range quotient {
		if diff < 0 {

			c.outBuf.WriteByte('-')
		} else {
			c.outBuf.WriteByte('+')

		}
	}
	//point to counter cell to decrement it
	for range c.currCell {
		c.outBuf.WriteByte('<')
	}
	//this will set currLoopCounter to 0 effectively
	c.currLoopCounter = 0
	c.outBuf.Write([]byte{'-', ']'})
	//return to currCell
	for range c.currCell {
		c.outBuf.WriteByte('>')
	}
	if diffWasIncremented {
		if diff < 0 {
			c.outBuf.WriteByte('+')
		} else {
			c.outBuf.WriteByte('-')
		}
	}
}
func absInt(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}
func average(i []int) int {
	sum := 0
	for _, v := range i {
		sum += v
	}
	return int(sum / len(i))
}
func (c *Converter) prepareInitialMem(v []int) {
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
		c.currLoopCounter++
		c.outBuf.WriteByte('+')
	}
	//dont forget to increment and decrement current cell
	c.outBuf.Write([]byte{'[', '>'})
	c.currCell++
	quotient := avgVal / divisor
	for range numOfCells {
		for range quotient {
			c.outBuf.WriteByte('+')
		}
		c.outBuf.WriteByte('>')
		c.currCell++
	}
	for c.currCell > 0 {
		c.outBuf.WriteByte('<')
		c.currCell -= 1
	}
	//adding - to the end of the loop decreases currLoopCounter to 0 in fact
	c.currLoopCounter = 0
	c.outBuf.Write([]byte{'-', ']'})
	for i, val := range v {
		diff := val - avgVal
		if diff != 0 {
			c.currCell++
			c.outBuf.WriteByte('>')
			//loop can be huge if we need to go long way to the counter; 3 is for [ ] and -
			if absInt(diff) >= 12 && absInt(diff) > c.currCell*2+3 {
				c.incrementCurrentCellBy(diff)
			} else {
				for range absInt(diff) {
					if diff > 0 {
						c.outBuf.WriteByte('-')
					} else {
						c.outBuf.WriteByte('+')
					}
				}
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
	f, err := os.Create("bfout.txt")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer f.Close()
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
	//fmt.Println(charslc)
	conv.prepareInitialMem(charslc)
	conv.preparePrintingPart(input)
	conv.output(os.Stdout)
	conv.output(f)
}
