/*
Copyright © 2021 Zoraiz Hassan <hzoraiz8@gmail.com>

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

package image_conversions

var (
	// Reference taken from http://paulbourke.net/dataformats/asciiart/
	asciiTableSimple   = " .:-=+*#%@"
	asciiTableDetailed = " .'`^\",:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"

	// Structure for braille dots
	BrailleStruct = [4][2]int{
		{0x1, 0x8},
		{0x2, 0x10},
		{0x4, 0x20},
		{0x40, 0x80},
	}

	BrailleThreshold uint32
)

// For each individual element of imgSet in ConvertToASCIISlice()
const MAX_VAL float64 = 255

type AsciiChar struct {
	OriginalColor string
	SetColor      string
	Simple        string
	RgbValue      [3]uint32
}

/*
Converts the 2D image_conversions.AsciiPixel slice of image data (each instance representing each compressed pixel of original image)
to a 2D image_conversions.AsciiChar slice

If complex parameter is true, values are compared to 70 levels of color density in ASCII characters.
Otherwise, values are compared to 10 levels of color density in ASCII characters.
*/
func ConvertToAsciiChars(imgSet [][]AsciiPixel, negative, colored, grayscale, complex, colorBg bool, customMap string, fontColor [3]int) ([][]AsciiChar, error) {

	height := len(imgSet)
	width := len(imgSet[0])

	chosenTable := map[int]string{}

	// Turn ascii character-set string into map[int]string{} literal
	if customMap == "" {
		var charSet string

		if complex {
			charSet = asciiTableDetailed
		} else {
			charSet = asciiTableSimple
		}

		for index, char := range charSet {
			chosenTable[index] = string(char)
		}

	} else {
		chosenTable = map[int]string{}

		for index, char := range customMap {
			chosenTable[index] = string(char)
		}
	}

	var result [][]AsciiChar

	for i := 0; i < height; i++ {

		var tempSlice []AsciiChar

		for j := 0; j < width; j++ {
			value := float64(imgSet[i][j].charDepth)

			// Gets appropriate string index from chosenTable by percentage comparisons with its length
			tempFloat := (value / MAX_VAL) * float64(len(chosenTable))
			if value == MAX_VAL {
				tempFloat = float64(len(chosenTable) - 1)
			}
			tempInt := int(tempFloat)

			var r, g, b int

			if colored {
				r = int(imgSet[i][j].rgbValue[0])
				g = int(imgSet[i][j].rgbValue[1])
				b = int(imgSet[i][j].rgbValue[2])
			} else {
				r = int(imgSet[i][j].grayscaleValue[0])
				g = int(imgSet[i][j].grayscaleValue[1])
				b = int(imgSet[i][j].grayscaleValue[2])
			}

			if negative {
				// Select character from opposite side of table as well as turn pixels negative
				r = 255 - r
				g = 255 - g
				b = 255 - b

				// To preserve negative rgb values for saving png image later down the line, since it uses imgSet
				if colored {
					imgSet[i][j].rgbValue = [3]uint32{uint32(r), uint32(g), uint32(b)}
				} else {
					imgSet[i][j].grayscaleValue = [3]uint32{uint32(r), uint32(g), uint32(b)}
				}

				tempInt = (len(chosenTable) - 1) - tempInt
			}

			var char AsciiChar

			asciiChar := chosenTable[tempInt]
			char.Simple = asciiChar

			var err error
			if colorBg {
				char.OriginalColor, err = getColoredCharForTerm(uint8(r), uint8(g), uint8(b), asciiChar, true)
			} else {
				char.OriginalColor, err = getColoredCharForTerm(uint8(r), uint8(g), uint8(b), asciiChar, false)
			}
			if (colored || grayscale) && err != nil {
				return nil, err
			}

			// If font color is not set, use a simple string. Otherwise, use True color
			if fontColor != [3]int{255, 255, 255} {
				fcR := fontColor[0]
				fcG := fontColor[1]
				fcB := fontColor[2]

				if colorBg {
					char.SetColor, err = getColoredCharForTerm(uint8(fcR), uint8(fcG), uint8(fcB), asciiChar, true)
				} else {
					char.SetColor, err = getColoredCharForTerm(uint8(fcR), uint8(fcG), uint8(fcB), asciiChar, false)
				}
				if err != nil {
					return nil, err
				}
			}

			if colored {
				char.RgbValue = imgSet[i][j].rgbValue
			} else {
				char.RgbValue = imgSet[i][j].grayscaleValue
			}

			tempSlice = append(tempSlice, char)
		}
		result = append(result, tempSlice)
	}

	return result, nil
}

func getRGB(pixel AsciiPixel, negative, colored bool) (int, int, int) {
	var r, g, b int

	if colored {
		r = int(pixel.rgbValue[0])
		g = int(pixel.rgbValue[1])
		b = int(pixel.rgbValue[2])
	} else {
		r = int(pixel.grayscaleValue[0])
		g = int(pixel.grayscaleValue[1])
		b = int(pixel.grayscaleValue[2])
	}

	if negative {
		// Select character from opposite side of table as well as turn pixels negative
		r = 255 - r
		g = 255 - g
		b = 255 - b

		// To preserve negative rgb values for saving png image later down the line, since it uses imgSet
		if colored {
			pixel.rgbValue = [3]uint32{uint32(r), uint32(g), uint32(b)}
		} else {
			pixel.grayscaleValue = [3]uint32{uint32(r), uint32(g), uint32(b)}
		}
	}
	return r, g, b
}

func ConvertToHalfBlockChars(imgSet [][]AsciiPixel, negative, colored, grayscale bool) ([][]AsciiChar, error) {

	halfBlock := "▀"
	height := len(imgSet)
	width := len(imgSet[0])

	var result [][]AsciiChar

	for i := 0; i < height/2; i++ {

		var tempSlice []AsciiChar
		var (
			last_upper_r, last_upper_g, last_upper_b int = -1, -1, -1
			last_lower_r, last_lower_g, last_lower_b int = -1, -1, -1
		)

		var sameColorChars AsciiChar
		for j := 0; j < width; j++ {

			upper_r, upper_g, upper_b := getRGB(imgSet[2*i][j], negative, colored)
			lower_r, lower_g, lower_b := 255, 255, 255
			if 2*i+1 < height {
				lower_r, lower_g, lower_b = getRGB(imgSet[2*i+1][j], negative, colored)
			}
			sameColorChars.Simple += halfBlock
			if upper_b == last_upper_b &&
				upper_g == last_upper_g &&
				upper_r == last_upper_r &&
				lower_b == last_lower_b &&
				lower_g == last_lower_g &&
				lower_r == last_lower_r &&
				j < width {

				continue
			}

			charFg, err := getColoredCharForTerm(uint8(upper_r), uint8(upper_g), uint8(upper_b), sameColorChars.Simple, false)
			charFgAndBg, err := getColoredCharForTerm(uint8(lower_r), uint8(lower_g), uint8(lower_b), charFg, true)
			if (colored || grayscale) && err != nil {
				return nil, err
			}

			sameColorChars.OriginalColor = charFgAndBg
			if colored {
				sameColorChars.RgbValue = imgSet[i][j].rgbValue
			} else {
				sameColorChars.RgbValue = imgSet[i][j].grayscaleValue
			}

			tempSlice = append(tempSlice, sameColorChars)

			sameColorChars.Simple = ""
			last_upper_b = upper_b
			last_upper_g = upper_g
			last_upper_r = upper_r
			last_lower_b = lower_b
			last_lower_g = lower_g
			last_lower_r = lower_r
		}
		result = append(result, tempSlice)
	}

	return result, nil
}

/*
Converts the 2D image_conversions.AsciiPixel slice of image data (each instance representing each compressed pixel of original image)
to a 2D image_conversions.AsciiChar slice

Unlike ConvertToAsciiChars(), this function calculates braille characters instead of ascii
*/
func ConvertToBrailleChars(imgSet [][]AsciiPixel, negative, colored, grayscale, colorBg bool, fontColor [3]int, threshold int) ([][]AsciiChar, error) {

	BrailleThreshold = uint32(threshold)

	height := len(imgSet)
	width := len(imgSet[0])

	var result [][]AsciiChar

	for i := 0; i < height; i += 4 {

		var tempSlice []AsciiChar

		var (
			last_r, last_g, last_b int = -1, -1, -1
		)
		var char AsciiChar

		for j := 0; j < width; j += 2 {

			brailleChar := getBrailleChar(i, j, negative, imgSet)
			r, g, b := getRGB(imgSet[i][j], negative, colored)
			char.Simple += brailleChar

			if last_b == b &&
				last_r == r &&
				last_g == g &&
				j+2 < width {
				continue
			}

			var err error
			if colorBg {
				char.OriginalColor, err = getColoredCharForTerm(uint8(r), uint8(g), uint8(b), char.Simple, true)
			} else {
				char.OriginalColor, err = getColoredCharForTerm(uint8(r), uint8(g), uint8(b), char.Simple, false)
			}
			if (colored || grayscale) && err != nil {
				return nil, err
			}

			// If font color is not set, use a simple string. Otherwise, use True color
			if fontColor != [3]int{255, 255, 255} {
				fcR := fontColor[0]
				fcG := fontColor[1]
				fcB := fontColor[2]

				if colorBg {
					char.SetColor, err = getColoredCharForTerm(uint8(fcR), uint8(fcG), uint8(fcB), char.Simple, true)
				} else {
					char.SetColor, err = getColoredCharForTerm(uint8(fcR), uint8(fcG), uint8(fcB), char.Simple, false)
				}
				if err != nil {
					return nil, err
				}
			}

			if colored {
				char.RgbValue = imgSet[i][j].rgbValue
			} else {
				char.RgbValue = imgSet[i][j].grayscaleValue
			}

			tempSlice = append(tempSlice, char)
			last_r, last_g, last_b = r, g, b
			char.Simple = ""
		}

		result = append(result, tempSlice)
	}

	return result, nil
}

// Iterate through the BrailleStruct table to see which dots need to be highlighted
func getBrailleChar(x, y int, negative bool, imgSet [][]AsciiPixel) string {

	brailleChar := 0x2800

	for i := 0; i < 4; i++ {
		for j := 0; j < 2; j++ {
			if negative {
				if imgSet[x+i][y+j].charDepth <= BrailleThreshold {
					brailleChar += BrailleStruct[i][j]
				}
			} else {
				if imgSet[x+i][y+j].charDepth >= BrailleThreshold {
					brailleChar += BrailleStruct[i][j]
				}
			}
		}
	}

	return string(brailleChar)
}
