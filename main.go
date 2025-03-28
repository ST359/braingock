//go:build js && wasm

package main

import (
	"bytes"
	"math"
	"syscall/js"
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
func TranslateToBf(in string) string {
	conv := new(Converter)
	conv.charset = make(map[int]struct{})
	conv.mem = make(map[int]int)
	charslc := make([]int, 0)
	for _, v := range []rune(in) {
		if _, ok := conv.charset[int(v)]; !ok {
			charslc = append(charslc, int(v))
			conv.charset[int(v)] = struct{}{}
		}
	}
	conv.prepareInitialMem2(charslc)
	conv.preparePrintingPart(in)
	res := conv.outBuf.String()
	return res
}

func jsWrapper() js.Func {
	resFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		in := args[0].String()
		translated := TranslateToBf(in)
		return translated
	})
	return resFunc
}

type Converter struct {
	currCell        int
	currLoopCounter int
	mem             map[int]int
	charset         map[int]struct{}
	outBuf          bytes.Buffer
}

func (c *Converter) incrementCurrentCellBy(diff int) {

	absDiff := absInt(diff)
	divisor := calculateElegantDivisor(absDiff)
	diffWasIncremented := false
	if divisor == 1 {
		absDiff++
		divisor = calculateElegantDivisor(absDiff)
		diffWasIncremented = true
	}
	quotient := int(absDiff / divisor)
	//point to counter cell, however, if last action was to walk from pointer to cell, just remove x of >'s
	recentChars := c.outBuf.Bytes()[c.outBuf.Len()-c.currCell+1:]
	if string(recentChars) == string(bytes.Repeat([]byte{'>'}, c.currCell-1)) {
		c.outBuf.Truncate(c.outBuf.Len() - c.currCell + 1)
	} else {
		//return to counter cell which should be equal to 0
		c.outBuf.Write([]byte{'[', '<', ']'})
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
	//can optimize this to be [>]<<< if currCell is further from 1st(counter) cell than from last cell+3
	//[>] goto last+1 <<<go to currCell
	distanceToTheCellFromTheEnd := len(c.charset) + 2 - c.currCell //there are 2 cells at the beginning and we will arrive at empty
	if c.currCell-1 > distanceToTheCellFromTheEnd+3 {              //dist < and [>] symbols
		c.outBuf.Write([]byte{'[', '>', ']'})
		for range distanceToTheCellFromTheEnd {
			c.outBuf.WriteByte('<')
		}
	} else {
		for range c.currCell - 1 {
			c.outBuf.WriteByte('>')
		}
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
	if c.currCell > 4 {
		c.outBuf.Write([]byte{'[', '<', ']', '>'})
	} else {
		for range c.currCell - 1 {
			c.outBuf.WriteByte('<')
		}
	}
	//decrement loop counter, this will after loop set it to 0
	c.outBuf.WriteByte('-')
	c.currLoopCounter = 0
	c.outBuf.WriteByte(']')
	//in the end we are on counter cell
	if c.currCell-1 > distanceToTheCellFromTheEnd+4 { //dist < and >[>] symbols
		c.outBuf.Write([]byte{'>', '[', '>', ']'})
		for range distanceToTheCellFromTheEnd {
			c.outBuf.WriteByte('<')
		}
	} else {
		for range c.currCell - 1 {
			c.outBuf.WriteByte('>')
		}
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
func (c *Converter) prepareInitialMem2(v []int) {
	numOfCells := len(v)
	avgVal := average(v)
	//calculate the most elegant divisor for avgVal
	//it is needed to make number of "+" outside loop
	//roghly equal to + inside loop
	m := calculateElegantDivisor(avgVal)
	if m == 1 {
		avgVal++
		m = calculateElegantDivisor(avgVal)
	}
	n := avgVal / m
	//num of chars +2 to store m dynamically and leave anchor 0 cell
	for range numOfCells + 2 {
		c.outBuf.WriteByte('>')
		c.currCell++
	}
	//store num of chars
	for range numOfCells {
		c.outBuf.WriteByte('+')
	}
	//start outer loop
	c.outBuf.WriteByte('[')
	//go to the first empty space on the left and then one more to setup m
	//currCell will be dynamic, need to just evaluate to the len(v)+1 in the end
	c.outBuf.Write([]byte{'[', '<', ']', '<'})
	for range m {
		c.outBuf.WriteByte('+')
	}
	//loop for incrementing the cell
	c.outBuf.Write([]byte{'[', '>'})
	for range n {
		c.outBuf.WriteByte('+')
	}
	c.outBuf.Write([]byte{'<', '-', ']'})
	//go to fresh char(we stopped at m) and then to first empty, one to the left is numofchars
	c.outBuf.Write([]byte{'>', '[', '>', ']', '<', '-'})
	//close outer loop, we now are on cell which was set up for storing num of chars
	c.outBuf.WriteByte(']')
	c.currLoopCounter = 0
	//return to the 1st(loopCounter) cell, alternative is in the next section go < instead of >
	c.outBuf.Write([]byte{'<', '[', '<', ']'})
	c.currCell = 1
	for _, val := range v {
		diff := val - avgVal
		c.currCell++
		c.outBuf.WriteByte('>')
		if diff != 0 {
			//loop can be huge if we need to go long way to the counter; 3 is for [ ] and -
			if absInt(diff) >= 12 && absInt(diff) > c.currCell+3 {
				c.incrementCurrentCellBy(diff)
			} else {
				for range absInt(diff) {
					//if val is *higher* than average
					if diff > 0 {
						c.outBuf.WriteByte('+')
					} else {
						c.outBuf.WriteByte('-')
					}
				}
			}
		}
		c.mem[val] = c.currCell //0-th cell used for init loop
	}
}

// first version
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
	//at this moment currLoopCounter = 0 so [<] will bring us to the 2nd cell, [<]> to the first char
	//and [>]< to the last char
	for _, v := range input {
		target := c.mem[int(v)]
		distToTargetFromEnd := len(c.charset) + 2 - target
		distToTargetFromStart := target - 1 //we will arrive at 2nd cell by [<]
		diff := target - c.currCell
		absDiff := absInt(diff)
		if absDiff > distToTargetFromEnd+3 { //"[>]" chars
			c.outBuf.Write([]byte{'[', '>', ']'})
			for range distToTargetFromEnd {
				c.outBuf.WriteByte('<')
			}
			c.currCell = target
		} else if absDiff > distToTargetFromStart+3 { //"[<]" chars
			c.outBuf.Write([]byte{'[', '<', ']'})
			for range distToTargetFromStart {
				c.outBuf.WriteByte('>')
			}
			c.currCell = target
		} else {
			for range absDiff {
				if diff < 0 {
					c.currCell--
					c.outBuf.WriteByte('<')
				} else if diff > 0 {
					c.currCell++
					c.outBuf.WriteByte('>')
				}
			}
		}
		c.outBuf.WriteByte('.')
	}
}

func main() {
	// var input string
	// fmt.Println("Type something to convert into brainfuck code")
	// scanner := bufio.NewScanner(os.Stdin)
	// if scanner.Scan() {
	// 	input = scanner.Text()
	// }
	// conv := new(Converter)
	// conv.charset = make(map[int]struct{})
	// conv.mem = make(map[int]int)
	// for _, v := range input {
	// 	fmt.Print(int(v))
	// 	conv.charset[int(v)] = struct{}{}
	// }
	// charslc := make([]int, 0)
	// for v := range conv.charset {
	// 	charslc = append(charslc, v)
	// }
	// conv.prepareInitialMem2(charslc)
	// conv.preparePrintingPart(input)
	// conv.output(os.Stdout)
	js.Global().Set("TranslateToBf", jsWrapper())
	<-make(chan struct{})
}
