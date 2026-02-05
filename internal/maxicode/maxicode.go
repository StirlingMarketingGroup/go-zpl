// Package maxicode implements MaxiCode 2D barcode encoding.
// Forked from github.com/ingridhq/maxicode v1.4.0 (MIT license).
package maxicode

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/StirlingMarketingGroup/go-zpl/internal/maxicode/readsolomon"
)

// special characters
const (
	RS  = string(rune(30))
	GS  = string(rune(29))
	EOT = string(rune(4))
)

var maxiGrid = []int{
	/* ISO/IEC 16023 Figure 5 - MaxiCode Module Sequence */ /* 30 x 33 data grid */
	122, 121, 128, 127, 134, 133, 140, 139, 146, 145, 152, 151, 158, 157, 164, 163, 170, 169, 176, 175, 182, 181, 188, 187, 194, 193, 200, 199, 0, 0,
	124, 123, 130, 129, 136, 135, 142, 141, 148, 147, 154, 153, 160, 159, 166, 165, 172, 171, 178, 177, 184, 183, 190, 189, 196, 195, 202, 201, 817, 0,
	126, 125, 132, 131, 138, 137, 144, 143, 150, 149, 156, 155, 162, 161, 168, 167, 174, 173, 180, 179, 186, 185, 192, 191, 198, 197, 204, 203, 819, 818,
	284, 283, 278, 277, 272, 271, 266, 265, 260, 259, 254, 253, 248, 247, 242, 241, 236, 235, 230, 229, 224, 223, 218, 217, 212, 211, 206, 205, 820, 0,
	286, 285, 280, 279, 274, 273, 268, 267, 262, 261, 256, 255, 250, 249, 244, 243, 238, 237, 232, 231, 226, 225, 220, 219, 214, 213, 208, 207, 822, 821,
	288, 287, 282, 281, 276, 275, 270, 269, 264, 263, 258, 257, 252, 251, 246, 245, 240, 239, 234, 233, 228, 227, 222, 221, 216, 215, 210, 209, 823, 0,
	290, 289, 296, 295, 302, 301, 308, 307, 314, 313, 320, 319, 326, 325, 332, 331, 338, 337, 344, 343, 350, 349, 356, 355, 362, 361, 368, 367, 825, 824,
	292, 291, 298, 297, 304, 303, 310, 309, 316, 315, 322, 321, 328, 327, 334, 333, 340, 339, 346, 345, 352, 351, 358, 357, 364, 363, 370, 369, 826, 0,
	294, 293, 300, 299, 306, 305, 312, 311, 318, 317, 324, 323, 330, 329, 336, 335, 342, 341, 348, 347, 354, 353, 360, 359, 366, 365, 372, 371, 828, 827,
	410, 409, 404, 403, 398, 397, 392, 391, 80, 79, 0, 0, 14, 13, 38, 37, 3, 0, 45, 44, 110, 109, 386, 385, 380, 379, 374, 373, 829, 0,
	412, 411, 406, 405, 400, 399, 394, 393, 82, 81, 41, 0, 16, 15, 40, 39, 4, 0, 0, 46, 112, 111, 388, 387, 382, 381, 376, 375, 831, 830,
	414, 413, 408, 407, 402, 401, 396, 395, 84, 83, 42, 0, 0, 0, 0, 0, 6, 5, 48, 47, 114, 113, 390, 389, 384, 383, 378, 377, 832, 0,
	416, 415, 422, 421, 428, 427, 104, 103, 56, 55, 17, 0, 0, 0, 0, 0, 0, 0, 21, 20, 86, 85, 434, 433, 440, 439, 446, 445, 834, 833,
	418, 417, 424, 423, 430, 429, 106, 105, 58, 57, 0, 0, 0, 0, 0, 0, 0, 0, 23, 22, 88, 87, 436, 435, 442, 441, 448, 447, 835, 0,
	420, 419, 426, 425, 432, 431, 108, 107, 60, 59, 0, 0, 0, 0, 0, 0, 0, 0, 0, 24, 90, 89, 438, 437, 444, 443, 450, 449, 837, 836,
	482, 481, 476, 475, 470, 469, 49, 0, 31, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 54, 53, 464, 463, 458, 457, 452, 451, 838, 0,
	484, 483, 478, 477, 472, 471, 50, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 466, 465, 460, 459, 454, 453, 840, 839,
	486, 485, 480, 479, 474, 473, 52, 51, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 43, 468, 467, 462, 461, 456, 455, 841, 0,
	488, 487, 494, 493, 500, 499, 98, 97, 62, 61, 0, 0, 0, 0, 0, 0, 0, 0, 0, 27, 92, 91, 506, 505, 512, 511, 518, 517, 843, 842,
	490, 489, 496, 495, 502, 501, 100, 99, 64, 63, 0, 0, 0, 0, 0, 0, 0, 0, 29, 28, 94, 93, 508, 507, 514, 513, 520, 519, 844, 0,
	492, 491, 498, 497, 504, 503, 102, 101, 66, 65, 18, 0, 0, 0, 0, 0, 0, 0, 19, 30, 96, 95, 510, 509, 516, 515, 522, 521, 846, 845,
	560, 559, 554, 553, 548, 547, 542, 541, 74, 73, 33, 0, 0, 0, 0, 0, 0, 11, 68, 67, 116, 115, 536, 535, 530, 529, 524, 523, 847, 0,
	562, 561, 556, 555, 550, 549, 544, 543, 76, 75, 0, 0, 8, 7, 36, 35, 12, 0, 70, 69, 118, 117, 538, 537, 532, 531, 526, 525, 849, 848,
	564, 563, 558, 557, 552, 551, 546, 545, 78, 77, 0, 34, 10, 9, 26, 25, 0, 0, 72, 71, 120, 119, 540, 539, 534, 533, 528, 527, 850, 0,
	566, 565, 572, 571, 578, 577, 584, 583, 590, 589, 596, 595, 602, 601, 608, 607, 614, 613, 620, 619, 626, 625, 632, 631, 638, 637, 644, 643, 852, 851,
	568, 567, 574, 573, 580, 579, 586, 585, 592, 591, 598, 597, 604, 603, 610, 609, 616, 615, 622, 621, 628, 627, 634, 633, 640, 639, 646, 645, 853, 0,
	570, 569, 576, 575, 582, 581, 588, 587, 594, 593, 600, 599, 606, 605, 612, 611, 618, 617, 624, 623, 630, 629, 636, 635, 642, 641, 648, 647, 855, 854,
	728, 727, 722, 721, 716, 715, 710, 709, 704, 703, 698, 697, 692, 691, 686, 685, 680, 679, 674, 673, 668, 667, 662, 661, 656, 655, 650, 649, 856, 0,
	730, 729, 724, 723, 718, 717, 712, 711, 706, 705, 700, 699, 694, 693, 688, 687, 682, 681, 676, 675, 670, 669, 664, 663, 658, 657, 652, 651, 858, 857,
	732, 731, 726, 725, 720, 719, 714, 713, 708, 707, 702, 701, 696, 695, 690, 689, 684, 683, 678, 677, 672, 671, 666, 665, 660, 659, 654, 653, 859, 0,
	734, 733, 740, 739, 746, 745, 752, 751, 758, 757, 764, 763, 770, 769, 776, 775, 782, 781, 788, 787, 794, 793, 800, 799, 806, 805, 812, 811, 861, 860,
	736, 735, 742, 741, 748, 747, 754, 753, 760, 759, 766, 765, 772, 771, 778, 777, 784, 783, 790, 789, 796, 795, 802, 801, 808, 807, 814, 813, 862, 0,
	738, 737, 744, 743, 750, 749, 756, 755, 762, 761, 768, 767, 774, 773, 780, 779, 786, 785, 792, 791, 798, 797, 804, 803, 810, 809, 816, 815, 864, 863,
}

var maxiCodeSet = []int{
	/* from Appendix A - ASCII character to Code Set (e.g. 2 = Set B) */
	/* set 0 refers to special characters that fit into more than one set (e.g. GS) */
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 0, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 0, 0, 0, 5, 0, 2, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 0, 1, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 2,
	2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 4, 5, 5, 5, 5, 5, 5, 4, 5, 3, 4, 3, 5, 5, 4, 4, 3, 3, 3,
	4, 3, 5, 4, 4, 3, 3, 4, 3, 3, 3, 4, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
}

var maxiSymbolChar = []int{
	/* from Appendix A - ASCII character to symbol value */
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
	20, 21, 22, 23, 24, 25, 26, 30, 28, 29, 30, 35, 32, 53, 34, 35, 36, 37, 38, 39,
	40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 37,
	38, 39, 40, 41, 52, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 42, 43, 44, 45, 46, 0, 1, 2, 3,
	4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23,
	24, 25, 26, 32, 54, 34, 35, 36, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 47, 48,
	49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 36,
	37, 37, 38, 39, 40, 41, 42, 43, 38, 44, 37, 39, 38, 45, 46, 40, 41, 39, 40, 41,
	42, 42, 47, 43, 44, 43, 44, 45, 45, 46, 47, 46, 0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 32,
	33, 34, 35, 36, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 32, 33, 34, 35, 36,
}

var (
	rsEcc10 = readsolomon.NewEncoder(0x43, 10, 1)
	rsEcc20 = readsolomon.NewEncoder(0x43, 20, 1)
	rsEcc28 = readsolomon.NewEncoder(0x43, 28, 1)
)

// Encode encodes the given data into a MaxiCode symbol grid using the specified mode and ECI.
func Encode(mode, eci int, inputData string) (*SymbolGrid, error) {
	scmHeader := "[)>" + RS + "01" + GS

	if mode < 2 || mode > 6 {
		return nil, errors.New("only modes 2 to 6 supported")
	}

	codewords := make(maxiCodewords, 144)
	secondaryData := ""

	if mode == 2 || mode == 3 {
		if !strings.HasPrefix(inputData, scmHeader) {
			return nil, errors.New("invalid mode 2 / mode 3 structured carrier message header")
		}

		// Header + postcode + country code + service class + tracking code + SCAC + EOT (44 characters mode 2, 41 for mode 3).
		minLen := 41
		if mode == 2 {
			minLen = 44
		}

		if len(inputData) < minLen {
			return nil, errors.New("input data is shorter than mandatory UPS requirements")
		}

		if !strings.HasSuffix(inputData, EOT) {
			return nil, errors.New("input data should end with EOT marker")
		}

		// Extract the postcode, country code and service class and save in the primary data buffer.
		// Note: Group separators GS are removed.

		const postcodeStart = 9

		groups := strings.Split(inputData[postcodeStart:], GS)
		if len(groups) < 3 {
			return nil, errors.New("input data should contain postcode, country code and service class separated by Gs")
		}

		postcode := groups[0]
		countryCode := groups[1]
		serviceClass := groups[2]

		if l := len(postcode); (mode == 2 && l != 9) || (mode == 3 && l != 6) {
			return nil, errors.New("invalid postcode length")
		}

		if len(countryCode) != 3 {
			return nil, errors.New("invalid country code length")
		}

		if len(serviceClass) != 3 {
			return nil, errors.New("invalid service class length")
		}

		cc, err := strconv.Atoi(countryCode)
		if err != nil {
			return nil, errors.New("country code must be numeric")
		}

		sc, err := strconv.Atoi(serviceClass)
		if err != nil {
			return nil, errors.New("service class must be numeric")
		}

		if mode == 2 {
			// Note: upstream library enforces cc == 840 here, but real Zebra printers
			// accept any country code for Mode 2. Removed to match Zebra behavior.

			pc, err := strconv.Atoi(postcode)
			if err != nil {
				return nil, errors.New("postcode must be numeric in mode 2")
			}

			codewords.processPrimaryMode2(pc, cc, sc)
		} else {
			codewords.processPrimaryMode3(postcode, cc, sc)
		}

		// Copy the header and the 2 digit year into the secondary data buffer.
		secondaryData = inputData[:9]

		// Copy the rest of the data into the secondary data buffer.
		if len(groups) > 3 {
			secondaryData += strings.Join(groups[3:], GS)
		}
	} else {
		// Not using a primary, just copy all the data into the secondary data buffer.
		codewords[0] = mode
		secondaryData = inputData
	}

	if err := codewords.processSecondary(mode, eci, secondaryData); err != nil {
		return nil, err
	}

	// All the data is sorted - now do error correction.
	codewords.primaryDataCheck()

	eccLen := 40 // 84 data codewords, 40 error corrections
	if mode == 5 {
		eccLen = 56 // 68 data codewords, 56 error corrections
	}

	codewords.secondaryDataCheckEven(eccLen / 2)
	codewords.secondaryDataCheckOdd(eccLen / 2)

	var grid SymbolGrid

	// Copy data into symbol grid.
	for row := 0; row < 33; row++ {
		for column := 0; column < 30; column++ {
			g := maxiGrid[(row*30)+column] + 5
			block := g / 6
			bit := g % 6

			if block == 0 {
				continue
			}

			bitPattern := (codewords[block-1] & (32 >> bit)) >> (5 - bit)

			if bitPattern != 0 {
				grid.SetModule(row, column, true)
			}
		}
	}

	// Add orientation markings.
	grid.SetModule(0, 28, true) // Top right filler
	grid.SetModule(0, 29, true)
	grid.SetModule(9, 10, true) // Top left marker
	grid.SetModule(9, 11, true)
	grid.SetModule(10, 11, true)
	grid.SetModule(15, 7, true) // Left hand marker
	grid.SetModule(16, 8, true)
	grid.SetModule(16, 20, true) // Right hand marker
	grid.SetModule(17, 20, true)
	grid.SetModule(22, 10, true) // Bottom left marker
	grid.SetModule(23, 10, true)
	grid.SetModule(22, 17, true) // Bottom right marker
	grid.SetModule(23, 17, true)

	return &grid, nil
}

type maxiCodewords []int

func (codewords maxiCodewords) processPrimaryMode2(postcode, countryCode, serviceClass int) {
	const postcodeLen = 9

	codewords[0] = ((postcode & 0x03) << 4) | 2
	codewords[1] = (postcode & 0xfc) >> 2
	codewords[2] = (postcode & 0x3f00) >> 8
	codewords[3] = (postcode & 0xfc000) >> 14
	codewords[4] = (postcode & 0x3f00000) >> 20
	codewords[5] = ((postcode & 0x3c000000) >> 26) | ((postcodeLen & 0x3) << 4)
	codewords[6] = ((postcodeLen & 0x3c) >> 2) | ((countryCode & 0x3) << 4)
	codewords[7] = (countryCode & 0xfc) >> 2
	codewords[8] = ((countryCode & 0x300) >> 8) | ((serviceClass & 0xf) << 2)
	codewords[9] = (serviceClass & 0x3f0) >> 4
}

func (codewords maxiCodewords) processPrimaryMode3(postcode string, countryCode, serviceClass int) {
	const postcodeLen = 6

	pc := make([]int, postcodeLen)

	for i := 0; i < postcodeLen; i++ {
		c := postcode[i]

		switch {
		// (Capital) letters shifted to Code Set A values.
		case c >= 'A' && c <= 'Z':
			pc[i] = int(c - 64)
		// Not a valid postcode character.
		case c == 27 || c == 31 || c == 33 || c >= 59:
			pc[i] = int(' ')
		// Input characters lower than 27 (NUL - SUB) in postcode are
		// interpreted as capital letters in Code Set A (e.g. LF becomes 'J').
		default:
			pc[i] = int(c)
		}
	}

	codewords[0] = ((pc[5] & 0x03) << 4) | 3
	codewords[1] = ((pc[4] & 0x03) << 4) | ((pc[5] & 0x3c) >> 2)
	codewords[2] = ((pc[3] & 0x03) << 4) | ((pc[4] & 0x3c) >> 2)
	codewords[3] = ((pc[2] & 0x03) << 4) | ((pc[3] & 0x3c) >> 2)
	codewords[4] = ((pc[1] & 0x03) << 4) | ((pc[2] & 0x3c) >> 2)
	codewords[5] = ((pc[0] & 0x03) << 4) | ((pc[1] & 0x3c) >> 2)
	codewords[6] = ((pc[0] & 0x3c) >> 2) | ((countryCode & 0x3) << 4)
	codewords[7] = (countryCode & 0xfc) >> 2
	codewords[8] = ((countryCode & 0x300) >> 8) | ((serviceClass & 0xf) << 2)
	codewords[9] = (serviceClass & 0x3f0) >> 4
}

func determineSet(set []int, i int, allowedSets ...int) int {
	opt1 := set[i-1]
	opt2 := set[i+1]

	opt1Allowed := slices.Index(allowedSets, opt1) != -1
	opt2Allowed := slices.Index(allowedSets, opt2) != -1

	if opt1Allowed && opt2Allowed {
		return min(opt1, opt2)
	}

	if opt1Allowed {
		return opt1
	}

	if opt2Allowed {
		return opt2
	}

	return allowedSets[0]
}

func (codewords maxiCodewords) processSecondary(mode, eci int, secondaryData string) error {
	// Format text according to Appendix A

	if secondaryData == "" {
		return errors.New("input data is empty")
	}

	if len(secondaryData) > 138 {
		return errors.New("input data is too long")
	}

	set := make([]int, 144)
	character := make([]int, 144)

	for i := 0; i < len(secondaryData); i++ {
		// Look up characters in table from Appendix A - this gives value and code set for most characters.
		set[i] = maxiCodeSet[secondaryData[i]]
		character[i] = maxiSymbolChar[secondaryData[i]]
	}

	// If a character can be represented in more than one code set, pick which version to use.
	if set[0] == 0 {
		if character[0] == 13 {
			character[0] = 0
		}

		set[0] = 1
	}

	dataLen := len(secondaryData)

	for i := 1; i < dataLen; i++ {
		// special characters
		if set[i] == 0 {
			switch character[i] {
			// Carriage Return
			case 13:
				set[i] = determineSet(set, i, 1, 5)
				if set[i] == 5 {
					character[i] = 13
				} else {
					character[i] = 0
				}
			// FS
			case 28:
				set[i] = determineSet(set, i, 1, 2, 3, 4, 5)
				if set[i] == 5 {
					character[i] = 32
				}
			// GS
			case 29:
				set[i] = determineSet(set, i, 1, 2, 3, 4, 5)
				if set[i] == 5 {
					character[i] = 33
				}
			// RS
			case 30:
				set[i] = determineSet(set, i, 1, 2, 3, 4, 5)
				if set[i] == 5 {
					character[i] = 34
				}
			// Space
			case 32:
				set[i] = determineSet(set, i, 1, 2, 3, 4, 5)

				switch set[i] {
				case 1:
					character[i] = 32
				case 2:
					character[i] = 47
				default:
					character[i] = 59
				}
			// Comma
			case 44:
				set[i] = determineSet(set, i, 1, 2)
				if set[i] == 2 {
					character[i] = 48
				}
			// Full Stop
			case 46:
				set[i] = determineSet(set, i, 1, 2)
				if set[i] == 2 {
					character[i] = 49
				}
			// Slash
			case 47:
				set[i] = determineSet(set, i, 1, 2)
				if set[i] == 2 {
					character[i] = 50
				}
			// Colon
			case 58:
				set[i] = determineSet(set, i, 1, 2)
				if set[i] == 2 {
					character[i] = 51
				}
			}
		}
	}

	// Add the padding
	for i := dataLen; i < 144; i++ {
		if set[dataLen-1] == 2 {
			set[i] = 2
		} else {
			set[i] = 1
		}

		character[i] = 33
	}

	// Find candidates for number compression.
	// Number compression not allowed on the primary data.
	// In modes 4,5 & 6 the first nine characters of the secondary data
	// are placed in the primary data.
	start := 9
	if mode == 2 || mode == 3 {
		start = 0
	}

	count := 0
	for i := start; i < 144; i++ {
		// Character is a number
		if set[i] == 1 && character[i] >= 48 && character[i] <= 57 {
			count++
		} else {
			count = 0
		}

		// Nine digits in a row can be compressed
		if count == 9 {
			for j := i - 8; j <= i; j++ {
				set[j] = 6
			}

			count = 0
		}
	}

	// Add shift and latch characters
	currSet := 1
	idx := 0

	for {
		if (set[idx] != currSet) && (set[idx] != 6) {
			switch set[idx] {
			case 1:
				if currSet == 2 {
					// Set B
					if idx+1 < 144 && set[idx+1] == 1 {
						if idx+2 < 144 && set[idx+2] == 1 {
							if idx+3 < 144 && set[idx+3] == 1 {
								// Latch A
								insertPosition(set, character, idx, &dataLen)
								character[idx] = 63 // Set B Latch A
								currSet = 1
								idx += 3 // Next 3 Set A so skip over
							} else {
								// 3 Shift A
								insertPosition(set, character, idx, &dataLen)
								character[idx] = 57 // Set B triple shift A
								idx += 2            // Next 2 Set A so skip over
							}
						} else {
							// 2 Shift A
							insertPosition(set, character, idx, &dataLen)
							character[idx] = 56 // Set B double shift A
							idx++               // Next Set A so skip over
						}
					} else {
						// Shift A
						insertPosition(set, character, idx, &dataLen)
						character[idx] = 59 // Set A Shift B
					}
				} else {
					// All sets other than B only have latch
					// Latch A
					insertPosition(set, character, idx, &dataLen)
					character[idx] = 58 // Sets C,D,E Latch A
					currSet = 1
				}
			case 2:
				// Set B
				// If not Set A or next Set B
				if currSet != 1 || (idx+1 < 144 && set[idx+1] == 2) {
					// Latch B
					insertPosition(set, character, idx, &dataLen)
					character[idx] = 63 // Sets A,C,D,E Latch B
					currSet = 2
				} else {
					// Only available from Set A
					// Shift B
					insertPosition(set, character, idx, &dataLen)
					character[idx] = 59 // Set B Shift A
				}
			case 3, 4, 5:
				// Set C, D, E
				// If first and next 3 same set, or not first and previous and next 2 same set
				if (idx == 0 && idx+3 < 144 && set[idx+1] == set[idx] && set[idx+2] == set[idx] && set[idx+3] == set[idx]) || (idx > 0 && set[idx-1] == set[idx] && idx+2 < 144 && set[idx+1] == set[idx] && set[idx+2] == set[idx]) {
					if idx == 0 {
						// Lock in C/D/E
						insertPosition(set, character, idx, &dataLen)
						character[idx] = 60 + set[idx] - 3
						idx++ // Extra bump
						insertPosition(set, character, idx, &dataLen)
						character[idx] = 60 + set[idx] - 3
						idx += 3 // Next 3 same set so skip over
					} else {
						// Add single Shift to previous Shift
						insertPosition(set, character, idx, &dataLen)
						character[idx-1] = 60 + set[idx] - 3
						idx += 2 // Next 2 same set so skip over
					}
					currSet = set[idx]
				} else {
					// Shift C/D/E
					insertPosition(set, character, idx, &dataLen)
					character[idx] = 60 + set[idx] - 3
				}
			}
			idx++ // Allow for bump
		}
		idx++

		if idx >= 144 {
			break
		}
	}

	// Number compression
	idx = 0
	for {
		if set[idx] == 6 {
			compressed := ""
			for j := 0; j < 9; j++ {
				compressed += string(rune(character[idx+j]))
			}

			value, err := strconv.Atoi(compressed)
			if err != nil {
				return fmt.Errorf("failed to convert compressed number: %w", err)
			}

			character[idx] = 31 // NS
			character[idx+1] = (value & 0x3f000000) >> 24
			character[idx+2] = (value & 0xfc0000) >> 18
			character[idx+3] = (value & 0x3f000) >> 12
			character[idx+4] = (value & 0xfc0) >> 6
			character[idx+5] = value & 0x3f

			idx += 6
			for j := idx; j < 140; j++ {
				set[j] = set[j+3]
				character[j] = character[j+3]
			}

			dataLen -= 3
		} else {
			idx++
		}

		if idx >= 144 {
			break
		}
	}

	// Insert ECI at the beginning of message if needed.
	// Encode ECI assignment numbers according to table 3.
	if eci != 0 {
		insertPosition(set, character, 0, &dataLen)
		character[0] = 27 // ECI

		switch {
		case eci <= 31:
			insertPosition(set, character, 1, &dataLen)
			character[1] = eci

		case eci >= 32 && eci <= 1023:
			insertPosition(set, character, 1, &dataLen)
			insertPosition(set, character, 1, &dataLen)
			insertPosition(set, character, 1, &dataLen)
			character[1] = 0x30 + ((eci >> 12) & 0x03)
			character[2] = (eci >> 6) & 0x3f
			character[3] = eci & 0x3f

		case eci >= 32768:
			insertPosition(set, character, 1, &dataLen)
			insertPosition(set, character, 1, &dataLen)
			insertPosition(set, character, 1, &dataLen)
			insertPosition(set, character, 1, &dataLen)
			character[1] = 0x38 + ((eci >> 18) & 0x02)
			character[2] = (eci >> 12) & 0x3f
			character[3] = (eci >> 6) & 0x3f
			character[4] = eci & 0x3f
		}
	}

	switch mode {
	case 2, 3:
		if dataLen > 84 {
			return errors.New("input data is too long")
		}

		for i := 0; i < 84; i++ {
			codewords[i+20] = character[i]
		}
	case 4, 6:
		if dataLen > 93 {
			return errors.New("input data is too long")
		}

		// Primary
		for i := 0; i < 9; i++ {
			codewords[i+1] = character[i]
		}

		// Secondary
		for i := 0; i < 84; i++ {
			codewords[i+20] = character[i+9]
		}
	case 5:
		if dataLen > 77 {
			return errors.New("input data is too long")
		}

		// Primary
		for i := 0; i < 9; i++ {
			codewords[i+1] = character[i]
		}

		// Secondary
		for i := 0; i < 68; i++ {
			codewords[i+20] = character[i+9]
		}
	}

	return nil
}

func (codewords maxiCodewords) primaryDataCheck() {
	const (
		dataLen = 10
		eccLen  = 10
	)

	data := make([]byte, dataLen)
	result := make([]byte, eccLen)

	for j := 0; j < dataLen; j++ {
		data[j] = byte(codewords[j])
	}

	getRsEncoder(eccLen).Encode(dataLen, data, result)

	for j := 0; j < eccLen; j++ {
		codewords[dataLen+j] = int(result[eccLen-1-j])
	}
}

func (codewords maxiCodewords) secondaryDataCheckEven(eccLen int) {
	dataLen := 68
	if eccLen == 20 {
		dataLen = 84
	}

	data := make([]byte, dataLen)
	result := make([]byte, eccLen)

	for j := 0; j < dataLen+1; j += 2 {
		data[j/2] = byte(codewords[j+20])
	}

	getRsEncoder(eccLen).Encode(dataLen/2, data, result)

	for j := 0; j < eccLen; j++ {
		codewords[dataLen+(2*j)+20] = int(result[eccLen-1-j])
	}
}

func (codewords maxiCodewords) secondaryDataCheckOdd(eccLen int) {
	dataLen := 68
	if eccLen == 20 {
		dataLen = 84
	}

	data := make([]byte, dataLen)
	result := make([]byte, eccLen)

	for j := 1; j < dataLen+1; j += 2 {
		data[(j-1)/2] = byte(codewords[j+20])
	}

	getRsEncoder(eccLen).Encode(dataLen/2, data, result)

	for j := 0; j < eccLen; j++ {
		codewords[dataLen+(2*j)+1+20] = int(result[eccLen-1-j])
	}
}

func insertPosition(set, character []int, position int, dataLen *int) {
	for i := 143; i > position; i-- {
		set[i] = set[i-1]
		character[i] = character[i-1]
	}

	*dataLen++
}

func getRsEncoder(eccLen int) *readsolomon.Encoder {
	switch eccLen {
	case 10:
		return rsEcc10
	case 20:
		return rsEcc20
	default:
		return rsEcc28
	}
}
