package nrgrid

import (
	"fmt"
	"errors"
	"github.com/zhenggao2/ngapp/utils"
)

// DmrsSchInfo contains information of PDSCH/PUSCH DMRS per antenna port.
type DmrsSchInfo struct {
	CdmGroup int
	Delta    int
}

// refer to 3GPP 38.211 vf30
//  Table 7.4.1.1.2-1: Parameters for PDSCH DM-RS configuration type 1.
//  Table 6.4.1.1.3-1: Parameters for PUSCH DM-RS configuration type 1.
var DmrsSchCfgType1 = map[int]*DmrsSchInfo{
	0: {0, 0},
	1: {0, 0},
	2: {1, 1},
	3: {1, 1},
	4: {0, 0},
	5: {0, 0},
	6: {1, 1},
	7: {1, 1},
}

// refer to 3GPP 38.211 vf30
//  Table 7.4.1.1.2-2: Parameters for PDSCH DM-RS configuration type 2.
//  Table 6.4.1.1.3-2: Parameters for PUSCH DM-RS configuration type 2.
var DmrsSchCfgType2 = map[int]*DmrsSchInfo{
	0:  {0, 0},
	1:  {0, 0},
	2:  {1, 2},
	3:  {1, 2},
	4:  {2, 4},
	5:  {2, 4},
	6:  {0, 0},
	7:  {0, 0},
	8:  {1, 2},
	9:  {1, 2},
	10: {2, 4},
	11: {2, 4},
}

// refer to 3GPP 38.211 vh40
//  Table 7.4.1.1.2-3: PDSCH DM-RS positions l- for single-symbol DM-RS.
// key="td_mapping type_additional position"
//  key = '%s_%s_%s' % (td, tdraPdschMappingType, dmrsPdschAddPos)
var DmrsPdschPosOneSymb = map[string][]int{
	"2_typeA_pos0": nil, "2_typeA_pos1": nil, "2_typeA_pos2": nil, "2_typeA_pos3": nil,
	"3_typeA_pos0": {0}, "3_typeA_pos1": {0}, "3_typeA_pos2": {0}, "3_typeA_pos3": {0},
	"4_typeA_pos0": {0}, "4_typeA_pos1": {0}, "4_typeA_pos2": {0}, "4_typeA_pos3": {0},
	"5_typeA_pos0": {0}, "5_typeA_pos1": {0}, "5_typeA_pos2": {0}, "5_typeA_pos3": {0},
	"6_typeA_pos0": {0}, "6_typeA_pos1": {0}, "6_typeA_pos2": {0}, "6_typeA_pos3": {0},
	"7_typeA_pos0": {0}, "7_typeA_pos1": {0}, "7_typeA_pos2": {0}, "7_typeA_pos3": {0},
	"8_typeA_pos0": {0}, "8_typeA_pos1": {0, 7}, "8_typeA_pos2": {0, 7}, "8_typeA_pos3": {0, 7},
	"9_typeA_pos0": {0}, "9_typeA_pos1": {0, 7}, "9_typeA_pos2": {0, 7}, "9_typeA_pos3": {0, 7},
	"10_typeA_pos0": {0}, "10_typeA_pos1": {0, 9}, "10_typeA_pos2": {0, 6, 9}, "10_typeA_pos3": {0, 6, 9},
	"11_typeA_pos0": {0}, "11_typeA_pos1": {0, 9}, "11_typeA_pos2": {0, 6, 9}, "11_typeA_pos3": {0, 6, 9},
	"12_typeA_pos0": {0}, "12_typeA_pos1": {0, 9}, "12_typeA_pos2": {0, 6, 9}, "12_typeA_pos3": {0, 5, 8, 11},
	"13_typeA_pos0": {0}, "13_typeA_pos1": {0, 11}, "13_typeA_pos2": {0, 7, 11}, "13_typeA_pos3": {0, 5, 8, 11},
	"14_typeA_pos0": {0}, "14_typeA_pos1": {0, 11}, "14_typeA_pos2": {0, 7, 11}, "14_typeA_pos3": {0, 5, 8, 11},
	// 38.211 vh40
	"2_typeB_pos0" : {0}, "2_typeB_pos1" : {0}, "2_typeB_pos2" : {0}, "2_typeB_pos3" : {0},
	"3_typeB_pos0" : {0}, "3_typeB_pos1" : {0}, "3_typeB_pos2" : {0}, "3_typeB_pos3" : {0},
	"4_typeB_pos0" : {0}, "4_typeB_pos1" : {0}, "4_typeB_pos2" : {0}, "4_typeB_pos3" : {0},
	"5_typeB_pos0" : {0}, "5_typeB_pos1" : {0, 4}, "5_typeB_pos2" : {0, 4}, "5_typeB_pos3" : {0, 4},
	"6_typeB_pos0" : {0}, "6_typeB_pos1" : {0, 4}, "6_typeB_pos2" : {0, 4}, "6_typeB_pos3" : {0, 4},
	"7_typeB_pos0" : {0}, "7_typeB_pos1" : {0, 4}, "7_typeB_pos2" : {0, 4}, "7_typeB_pos3" : {0, 4},
	"8_typeB_pos0" : {0}, "8_typeB_pos1" : {0, 6}, "8_typeB_pos2" : {0, 3, 6}, "8_typeB_pos3" : {0, 3, 6},
	"9_typeB_pos0" : {0}, "9_typeB_pos1" : {0, 7}, "9_typeB_pos2" : {0, 4, 7}, "9_typeB_pos3" : {0, 4, 7},
	"10_typeB_pos0" : {0}, "10_typeB_pos1" : {0, 7}, "10_typeB_pos2" : {0, 4, 7}, "10_typeB_pos3" : {0, 4, 7},
	"11_typeB_pos0" : {0}, "11_typeB_pos1" : {0, 8}, "11_typeB_pos2" : {0, 4, 8}, "11_typeB_pos3" : {0, 3, 6, 9},
	"12_typeB_pos0" : {0}, "12_typeB_pos1" : {0, 9}, "12_typeB_pos2" : {0, 5, 9}, "12_typeB_pos3" : {0, 3, 6, 9},
	"13_typeB_pos0" : {0}, "13_typeB_pos1" : {0, 9}, "13_typeB_pos2" : {0, 5, 9}, "13_typeB_pos3" : {0, 3, 6, 9},
	"14_typeB_pos0" : nil, "14_typeB_pos1" : nil, "14_typeB_pos2" : nil, "14_typeB_pos3" : nil,
}

// refer to 3GPP 38.211 vh40
//  Table 7.4.1.1.2-4: PDSCH DM-RS positions l- for double-symbol DM-RS.
// key="td_mapping type_additional position"
//  key = '%s_%s_%s' % (td, tdraPdschMappingType, dmrsPdschAddPos)
var DmrsPdschPosTwoSymbs = map[string][]int{
	"2_typeA_pos0": nil, "2_typeA_pos1": nil,
	"3_typeA_pos0": nil, "3_typeA_pos1": nil,
	"4_typeA_pos0": {0}, "4_typeA_pos1": {0},
	"5_typeA_pos0": {0}, "5_typeA_pos1": {0},
	"6_typeA_pos0": {0}, "6_typeA_pos1": {0},
	"7_typeA_pos0": {0}, "7_typeA_pos1": {0},
	"8_typeA_pos0": {0}, "8_typeA_pos1": {0},
	"9_typeA_pos0": {0}, "9_typeA_pos1": {0},
	"10_typeA_pos0": {0}, "10_typeA_pos1": {0, 8},
	"11_typeA_pos0": {0}, "11_typeA_pos1": {0, 8},
	"12_typeA_pos0": {0}, "12_typeA_pos1": {0, 8},
	"13_typeA_pos0": {0}, "13_typeA_pos1": {0, 10},
	"14_typeA_pos0": {0}, "14_typeA_pos1": {0, 10},

	// new in vh40
	"2_typeB_pos0": nil, "2_typeB_pos1": nil,
	"3_typeB_pos0": nil, "3_typeB_pos1": nil,
	"4_typeB_pos0": nil, "4_typeB_pos1": nil,
	"5_typeB_pos0": {0}, "5_typeB_pos1": {0},
	"6_typeB_pos0": {0}, "6_typeB_pos1": {0},
	"7_typeB_pos0": {0}, "7_typeB_pos1": {0},
	"8_typeB_pos0": {0}, "8_typeB_pos1": {0, 5},
	"9_typeB_pos0": {0}, "9_typeB_pos1": {0, 5},
	"10_typeB_pos0": {0}, "10_typeB_pos1": {0, 7},
	"11_typeB_pos0": {0}, "11_typeB_pos1": {0, 7},
	"12_typeB_pos0": {0}, "12_typeB_pos1": {0, 8},
	"13_typeB_pos0": {0}, "13_typeB_pos1": {0, 8},
	"14_typeB_pos0": nil, "14_typeB_pos1": nil,
}

// refer to 3GPP 38.211 vh40
//  Table 6.4.1.1.3-3: PUSCH DM-RS positions l- within a slot for single-symbol DM-RS and intra-slot frequency hopping disabled.
// key="ld_mapping type_additional position"
//  key = '%s_%s_%s' % (ld, tdraPuschMappingType, dmrsPuschAddPos)
var DmrsPuschPosOneSymbWoIntraSlotFh = map[string][]int{
	"1_typeA_pos0": nil, "1_typeA_pos1": nil, "1_typeA_pos2": nil, "1_typeA_pos3": nil,
	"2_typeA_pos0": nil, "2_typeA_pos1": nil, "2_typeA_pos2": nil, "2_typeA_pos3": nil,
	"3_typeA_pos0": nil, "3_typeA_pos1": nil, "3_typeA_pos2": nil, "3_typeA_pos3": nil,
	"4_typeA_pos0": {0}, "4_typeA_pos1": {0}, "4_typeA_pos2": {0}, "4_typeA_pos3": {0},
	"5_typeA_pos0": {0}, "5_typeA_pos1": {0}, "5_typeA_pos2": {0}, "5_typeA_pos3": {0},
	"6_typeA_pos0": {0}, "6_typeA_pos1": {0}, "6_typeA_pos2": {0}, "6_typeA_pos3": {0},
	"7_typeA_pos0": {0}, "7_typeA_pos1": {0}, "7_typeA_pos2": {0}, "7_typeA_pos3": {0},
	"8_typeA_pos0": {0}, "8_typeA_pos1": {0, 7}, "8_typeA_pos2": {0, 7}, "8_typeA_pos3": {0, 7},
	"9_typeA_pos0": {0}, "9_typeA_pos1": {0, 7}, "9_typeA_pos2": {0, 7}, "9_typeA_pos3": {0, 7},
	"10_typeA_pos0": {0}, "10_typeA_pos1": {0, 9}, "10_typeA_pos2": {0, 6, 9}, "10_typeA_pos3": {0, 6, 9},
	"11_typeA_pos0": {0}, "11_typeA_pos1": {0, 9}, "11_typeA_pos2": {0, 6, 9}, "11_typeA_pos3": {0, 6, 9},
	"12_typeA_pos0": {0}, "12_typeA_pos1": {0, 9}, "12_typeA_pos2": {0, 6, 9}, "12_typeA_pos3": {0, 5, 8, 11},
	"13_typeA_pos0": {0}, "13_typeA_pos1": {0, 11}, "13_typeA_pos2": {0, 7, 11}, "13_typeA_pos3": {0, 5, 8, 11},
	"14_typeA_pos0": {0}, "14_typeA_pos1": {0, 11}, "14_typeA_pos2": {0, 7, 11}, "14_typeA_pos3": {0, 5, 8, 11},
	"1_typeB_pos0": {0}, "1_typeB_pos1": {0}, "1_typeB_pos2": {0}, "1_typeB_pos3": {0},
	"2_typeB_pos0": {0}, "2_typeB_pos1": {0}, "2_typeB_pos2": {0}, "2_typeB_pos3": {0},
	"3_typeB_pos0": {0}, "3_typeB_pos1": {0}, "3_typeB_pos2": {0}, "3_typeB_pos3": {0},
	"4_typeB_pos0": {0}, "4_typeB_pos1": {0}, "4_typeB_pos2": {0}, "4_typeB_pos3": {0},
	"5_typeB_pos0": {0}, "5_typeB_pos1": {0, 4}, "5_typeB_pos2": {0, 4}, "5_typeB_pos3": {0, 4},
	"6_typeB_pos0": {0}, "6_typeB_pos1": {0, 4}, "6_typeB_pos2": {0, 4}, "6_typeB_pos3": {0, 4},
	"7_typeB_pos0": {0}, "7_typeB_pos1": {0, 4}, "7_typeB_pos2": {0, 4}, "7_typeB_pos3": {0, 4},
	"8_typeB_pos0": {0}, "8_typeB_pos1": {0, 6}, "8_typeB_pos2": {0, 3, 6}, "8_typeB_pos3": {0, 3, 6},
	"9_typeB_pos0": {0}, "9_typeB_pos1": {0, 6}, "9_typeB_pos2": {0, 3, 6}, "9_typeB_pos3": {0, 3, 6},
	"10_typeB_pos0": {0}, "10_typeB_pos1": {0, 8}, "10_typeB_pos2": {0, 4, 8}, "10_typeB_pos3": {0, 3, 6, 9},
	"11_typeB_pos0": {0}, "11_typeB_pos1": {0, 8}, "11_typeB_pos2": {0, 4, 8}, "11_typeB_pos3": {0, 3, 6, 9},
	"12_typeB_pos0": {0}, "12_typeB_pos1": {0, 10}, "12_typeB_pos2": {0, 5, 10}, "12_typeB_pos3": {0, 3, 6, 9},
	"13_typeB_pos0": {0}, "13_typeB_pos1": {0, 10}, "13_typeB_pos2": {0, 5, 10}, "13_typeB_pos3": {0, 3, 6, 9},
	"14_typeB_pos0": {0}, "14_typeB_pos1": {0, 10}, "14_typeB_pos2": {0, 5, 10}, "14_typeB_pos3": {0, 3, 6, 9},
}

// refer to 3GPP 38.211 vh40
//  Table 6.4.1.1.3-4: PUSCH DM-RS positions l- within a slot for double-symbol DM-RS and intra-slot frequency hopping disabled.
// key="ld_mapping type_additional position"
//  key = '%s_%s_%s' % (ld, tdraPuschMappingType, dmrsPuschAddPos)
var DmrsPuschPosTwoSymbsWoIntraSlotFh = map[string][]int{
	"1_typeA_pos0": nil, "1_typeA_pos1": nil,
	"2_typeA_pos0": nil, "2_typeA_pos1": nil,
	"3_typeA_pos0": nil, "3_typeA_pos1": nil,
	"4_typeA_pos0": {0}, "4_typeA_pos1": {0},
	"5_typeA_pos0": {0}, "5_typeA_pos1": {0},
	"6_typeA_pos0": {0}, "6_typeA_pos1": {0},
	"7_typeA_pos0": {0}, "7_typeA_pos1": {0},
	"8_typeA_pos0": {0}, "8_typeA_pos1": {0},
	"9_typeA_pos0": {0}, "9_typeA_pos1": {0},
	"10_typeA_pos0": {0}, "10_typeA_pos1": {0, 8},
	"11_typeA_pos0": {0}, "11_typeA_pos1": {0, 8},
	"12_typeA_pos0": {0}, "12_typeA_pos1": {0, 8},
	"13_typeA_pos0": {0}, "13_typeA_pos1": {0, 10},
	"14_typeA_pos0": {0}, "14_typeA_pos1": {0, 10},
	"1_typeB_pos0": nil, "1_typeB_pos1": nil,
	"2_typeB_pos0": nil, "2_typeB_pos1": nil,
	"3_typeB_pos0": nil, "3_typeB_pos1": nil,
	"4_typeB_pos0": nil, "4_typeB_pos1": nil,
	"5_typeB_pos0": {0}, "5_typeB_pos1": {0},
	"6_typeB_pos0": {0}, "6_typeB_pos1": {0},
	"7_typeB_pos0": {0}, "7_typeB_pos1": {0},
	"8_typeB_pos0": {0}, "8_typeB_pos1": {0, 5},
	"9_typeB_pos0": {0}, "9_typeB_pos1": {0, 5},
	"10_typeB_pos0": {0}, "10_typeB_pos1": {0, 7},
	"11_typeB_pos0": {0}, "11_typeB_pos1": {0, 7},
	"12_typeB_pos0": {0}, "12_typeB_pos1": {0, 9},
	"13_typeB_pos0": {0}, "13_typeB_pos1": {0, 9},
	"14_typeB_pos0": {0}, "14_typeB_pos1": {0, 9},
}

// refer to 3GPP 38.211 vh40
//  Table 6.4.1.1.3-6: PUSCH DM-RS positions l- within a slot for single-symbol DM-RS and intra-slot frequency hopping enabled.
// key="ld per hop_mapping type_type a position_additional position_hop"
//  ld1 = math.floor(td / 2)
//  key1 = '%s_%s_%s_%s_1st' % (ld1, tdraPuschMappingType, dmrsTypeAPos if mappingType == 'typeA' else '0', 'pos1' if dmrsPuschAddPos != 'pos0' else 'pos0')
//  ld2 = td - math.floor(td / 2)
//  key2 = '%s_%s_%s_%s_2nd' % (ld2, tdraPuschMappingType, dmrsTypeAPos if mappingType == 'typeA' else '0', 'pos1' if dmrsPuschAddPos != 'pos0' else 'pos0')
var DmrsPuschPosOneSymbWithIntraSlotFh = map[string][]int{
	"1_typeA_2_pos0_1st": nil,
	"2_typeA_2_pos0_1st": nil,
	"3_typeA_2_pos0_1st": nil,
	"4_typeA_2_pos0_1st": {2},
	"5_typeA_2_pos0_1st": {2},
	"6_typeA_2_pos0_1st": {2},
	"7_typeA_2_pos0_1st": {2},
	"1_typeA_2_pos0_2nd": nil,
	"2_typeA_2_pos0_2nd": nil,
	"3_typeA_2_pos0_2nd": nil,
	"4_typeA_2_pos0_2nd": {0},
	"5_typeA_2_pos0_2nd": {0},
	"6_typeA_2_pos0_2nd": {0},
	"7_typeA_2_pos0_2nd": {0},
	"1_typeA_2_pos1_1st": nil,
	"2_typeA_2_pos1_1st": nil,
	"3_typeA_2_pos1_1st": nil,
	"4_typeA_2_pos1_1st": {2},
	"5_typeA_2_pos1_1st": {2},
	"6_typeA_2_pos1_1st": {2},
	"7_typeA_2_pos1_1st": {2, 6},
	"1_typeA_2_pos1_2nd": nil,
	"2_typeA_2_pos1_2nd": nil,
	"3_typeA_2_pos1_2nd": nil,
	"4_typeA_2_pos1_2nd": {0},
	"5_typeA_2_pos1_2nd": {0, 4},
	"6_typeA_2_pos1_2nd": {0, 4},
	"7_typeA_2_pos1_2nd": {0, 4},
	"1_typeA_3_pos0_1st": nil,
	"2_typeA_3_pos0_1st": nil,
	"3_typeA_3_pos0_1st": nil,
	"4_typeA_3_pos0_1st": {3},
	"5_typeA_3_pos0_1st": {3},
	"6_typeA_3_pos0_1st": {3},
	"7_typeA_3_pos0_1st": {3},
	"1_typeA_3_pos0_2nd": nil,
	"2_typeA_3_pos0_2nd": nil,
	"3_typeA_3_pos0_2nd": nil,
	"4_typeA_3_pos0_2nd": {0},
	"5_typeA_3_pos0_2nd": {0},
	"6_typeA_3_pos0_2nd": {0},
	"7_typeA_3_pos0_2nd": {0},
	"1_typeA_3_pos1_1st": nil,
	"2_typeA_3_pos1_1st": nil,
	"3_typeA_3_pos1_1st": nil,
	"4_typeA_3_pos1_1st": {3},
	"5_typeA_3_pos1_1st": {3},
	"6_typeA_3_pos1_1st": {3},
	"7_typeA_3_pos1_1st": {3},
	"1_typeA_3_pos1_2nd": nil,
	"2_typeA_3_pos1_2nd": nil,
	"3_typeA_3_pos1_2nd": nil,
	"4_typeA_3_pos1_2nd": {0},
	"5_typeA_3_pos1_2nd": {0, 4},
	"6_typeA_3_pos1_2nd": {0, 4},
	"7_typeA_3_pos1_2nd": {0, 4},
	"1_typeB_0_pos0_1st": {0},
	"2_typeB_0_pos0_1st": {0},
	"3_typeB_0_pos0_1st": {0},
	"4_typeB_0_pos0_1st": {0},
	"5_typeB_0_pos0_1st": {0},
	"6_typeB_0_pos0_1st": {0},
	"7_typeB_0_pos0_1st": {0},
	"1_typeB_0_pos0_2nd": {0},
	"2_typeB_0_pos0_2nd": {0},
	"3_typeB_0_pos0_2nd": {0},
	"4_typeB_0_pos0_2nd": {0},
	"5_typeB_0_pos0_2nd": {0},
	"6_typeB_0_pos0_2nd": {0},
	"7_typeB_0_pos0_2nd": {0},
	"1_typeB_0_pos1_1st": {0},
	"2_typeB_0_pos1_1st": {0},
	"3_typeB_0_pos1_1st": {0},
	"4_typeB_0_pos1_1st": {0},
	"5_typeB_0_pos1_1st": {0, 4},
	"6_typeB_0_pos1_1st": {0, 4},
	"7_typeB_0_pos1_1st": {0, 4},
	"1_typeB_0_pos1_2nd": {0},
	"2_typeB_0_pos1_2nd": {0},
	"3_typeB_0_pos1_2nd": {0},
	"4_typeB_0_pos1_2nd": {0},
	"5_typeB_0_pos1_2nd": {0, 4},
	"6_typeB_0_pos1_2nd": {0, 4},
	"7_typeB_0_pos1_2nd": {0, 4},
}

// refer to 3GPP 38.212 vh40
// note: 1st part of key: 0=fullyAndPartialAndNonCoherent, 1=partialAndNonCoherent, 2=nonCoherent
//  Table 7.3.1.1.2-2: Precoding information and number of layers, for 4 antenna ports, if transform precoder is disabled, maxRank = 2 or 3 or 4, and ul-FullPowerTransmission is not configured or configured to fullpowerMode2 or configured to fullpower
//  Table 7.3.1.1.2-2A: Precoding information and number of layers for 4 antenna ports, if transform precoder is disabled, maxRank = 2, and ul-FullPowerTransmission = fullpowerMode1
//  Table 7.3.1.1.2-2B: Precoding information and number of layers for 4 antenna ports, if transform precoder is disabled, maxRank = 3 or 4, and ul-FullPowerTransmission = fullpowerMode1
//  Table 7.3.1.1.2-2C: Second precoding information, for 4 antenna ports, if transform precoder is disabled, maxRank = 2 or 3 or 4, and ul-FullPowerTransmission is not configured or configured to fullpowerMode2 or configured to fullpower
//  Table 7.3.1.1.2-2D: Second precoding information for 4 antenna ports, if transform precoder is disabled, maxRank = 2, and ul-FullPowerTransmission = fullpowerMode1
//  Table 7.3.1.1.2-2E: Second precoding information for 4 antenna ports, if transform precoder is disabled, maxRank = 3 or 4, and ul-FullPowerTransmission = fullpowerMode1
var Dci01TpmiAp4Tp0MaxRank234 = map[string][]int{
	"0_0":  {1, 0},
	"0_1":  {1, 1},
	"0_2":  {1, 2},
	"0_3":  {1, 3},
	"0_4":  {2, 0},
	"0_5":  {2, 1},
	"0_6":  {2, 2},
	"0_7":  {2, 3},
	"0_8":  {2, 4},
	"0_9":  {2, 5},
	"0_10": {3, 0},
	"0_11": {4, 0},
	"0_12": {1, 4},
	"0_13": {1, 5},
	"0_14": {1, 6},
	"0_15": {1, 7},
	"0_16": {1, 8},
	"0_17": {1, 9},
	"0_18": {1, 10},
	"0_19": {1, 11},
	"0_20": {2, 6},
	"0_21": {2, 7},
	"0_22": {2, 8},
	"0_23": {2, 9},
	"0_24": {2, 10},
	"0_25": {2, 11},
	"0_26": {2, 12},
	"0_27": {2, 13},
	"0_28": {3, 1},
	"0_29": {3, 2},
	"0_30": {4, 1},
	"0_31": {4, 2},
	"0_32": {1, 12},
	"0_33": {1, 13},
	"0_34": {1, 14},
	"0_35": {1, 15},
	"0_36": {1, 16},
	"0_37": {1, 17},
	"0_38": {1, 18},
	"0_39": {1, 19},
	"0_40": {1, 20},
	"0_41": {1, 21},
	"0_42": {1, 22},
	"0_43": {1, 23},
	"0_44": {1, 24},
	"0_45": {1, 25},
	"0_46": {1, 26},
	"0_47": {1, 27},
	"0_48": {2, 14},
	"0_49": {2, 15},
	"0_50": {2, 16},
	"0_51": {2, 17},
	"0_52": {2, 18},
	"0_53": {2, 19},
	"0_54": {2, 20},
	"0_55": {2, 21},
	"0_56": {3, 3},
	"0_57": {3, 4},
	"0_58": {3, 5},
	"0_59": {3, 6},
	"0_60": {4, 3},
	"0_61": {4, 4},
	"0_62": nil,
	"0_63": nil,
	"1_0":  {1, 0},
	"1_1":  {1, 1},
	"1_2":  {1, 2},
	"1_3":  {1, 3},
	"1_4":  {2, 0},
	"1_5":  {2, 1},
	"1_6":  {2, 2},
	"1_7":  {2, 3},
	"1_8":  {2, 4},
	"1_9":  {2, 5},
	"1_10": {3, 0},
	"1_11": {4, 0},
	"1_12": {1, 4},
	"1_13": {1, 5},
	"1_14": {1, 6},
	"1_15": {1, 7},
	"1_16": {1, 8},
	"1_17": {1, 9},
	"1_18": {1, 10},
	"1_19": {1, 11},
	"1_20": {2, 6},
	"1_21": {2, 7},
	"1_22": {2, 8},
	"1_23": {2, 9},
	"1_24": {2, 10},
	"1_25": {2, 11},
	"1_26": {2, 12},
	"1_27": {2, 13},
	"1_28": {3, 1},
	"1_29": {3, 2},
	"1_30": {4, 1},
	"1_31": {4, 2},
	"2_0":  {1, 0},
	"2_1":  {1, 1},
	"2_2":  {1, 2},
	"2_3":  {1, 3},
	"2_4":  {2, 0},
	"2_5":  {2, 1},
	"2_6":  {2, 2},
	"2_7":  {2, 3},
	"2_8":  {2, 4},
	"2_9":  {2, 5},
	"2_10": {3, 0},
	"2_11": {4, 0},
	"2_12": nil,
	"2_13": nil,
	"2_14": nil,
}

// refer to 3GPP 38.212 vh40
// note: 1st part of key: 0=fullyAndPartialAndNonCoherent, 1=partialAndNonCoherent, 2=nonCoherent
//  Table 7.3.1.1.2-3: Precoding information and number of layers or Second Precoding information, for 4 antenna ports, if transform precoder is enabled and ul-FullPowerTransmission is either not configured or configured to fullpowerMode2 or configured to fullpower, or if transform precoder is disabled, maxRank = 1, and ul-FullPowerTransmission is not configured or configured to fullpowerMode2 or configured to fullpower
//  Table 7.3.1.1.2-3A: Precoding information and number of layers or Second Precoding information, for 4 antenna ports, if transform precoder is enabled and ul-FullPowerTransmission = fullpowerMode1, or if transform precoder is disabled, maxRank = 1, and ul-FullPowerTransmission = fullpowerMode1
var Dci01TpmiAp4Tp1OrTp0MaxRank1 = map[string][]int{
	"0_0":  {1, 0},
	"0_1":  {1, 1},
	"0_2":  {1, 2},
	"0_3":  {1, 3},
	"0_4":  {1, 4},
	"0_5":  {1, 5},
	"0_6":  {1, 6},
	"0_7":  {1, 7},
	"0_8":  {1, 8},
	"0_9":  {1, 9},
	"0_10": {1, 10},
	"0_11": {1, 11},
	"0_12": {1, 12},
	"0_13": {1, 13},
	"0_14": {1, 14},
	"0_15": {1, 15},
	"0_16": {1, 16},
	"0_17": {1, 17},
	"0_18": {1, 18},
	"0_19": {1, 19},
	"0_20": {1, 20},
	"0_21": {1, 21},
	"0_22": {1, 22},
	"0_23": {1, 23},
	"0_24": {1, 24},
	"0_25": {1, 25},
	"0_26": {1, 26},
	"0_27": {1, 27},
	"0_28": nil,
	"0_29": nil,
	"0_30": nil,
	"0_31": nil,
	"1_0":  {1, 0},
	"1_1":  {1, 1},
	"1_2":  {1, 2},
	"1_3":  {1, 3},
	"1_4":  {1, 4},
	"1_5":  {1, 5},
	"1_6":  {1, 6},
	"1_7":  {1, 7},
	"1_8":  {1, 8},
	"1_9":  {1, 9},
	"1_10": {1, 10},
	"1_11": {1, 11},
	"1_12": nil,
	"1_13": nil,
	"1_14": nil,
	"1_15": nil,
	"2_0":  {1, 0},
	"2_1":  {1, 1},
	"2_2":  {1, 2},
	"2_3":  {1, 3},
}

// refer to 3GPP 38.212 vh40
// note: 1st part of key: 0=fullyAndPartialAndNonCoherent, 1=partialAndNonCoherent, 2=nonCoherent
//  Table 7.3.1.1.2-4: Precoding information and number of layers, for 2 antenna ports, if transform precoder is disabled, maxRank = 2, and ul-FullPowerTransmission is not configured or configured to fullpowerMode2 or configured to fullpower
//  Table 7.3.1.1.2-4A: Precoding information and number of layers, for 2 antenna ports, if transform precoder is disabled, maxRank = 2, and ul-FullPowerTransmission = fullpowerMode1
//  Table 7.3.1.1.2-4B: Second precoding information, for 2 antenna ports, if transform precoder is disabled, maxRank = 2, and ul-FullPowerTransmission is not configured or configured to fullpowerMode2 or configured to fullpower
//  Table 7.3.1.1.2-4C: Second precoding information, for 2 antenna ports, if transform precoder is disabled, maxRank = 2, and ul-FullPowerTransmission = fullpowerMode1
var Dci01TpmiAp2Tp0MaxRank2 = map[string][]int{
	"0_0":  {1, 0},
	"0_1":  {1, 1},
	"0_2":  {2, 0},
	"0_3":  {1, 2},
	"0_4":  {1, 3},
	"0_5":  {1, 4},
	"0_6":  {1, 5},
	"0_7":  {2, 1},
	"0_8":  {2, 2},
	"0_9":  nil,
	"0_10": nil,
	"0_11": nil,
	"0_12": nil,
	"0_13": nil,
	"0_14": nil,
	"0_15": nil,
	"2_0":  {1, 0},
	"2_1":  {1, 1},
	"2_2":  {2, 0},
	"2_3":  nil,
}

// refer to 3GPP 38.212 vh40
// note: 1st part of key: 0=fullyAndPartialAndNonCoherent, 1=partialAndNonCoherent, 2=nonCoherent
//  Table 7.3.1.1.2-5: Precoding information and number of layers or Second Precoding information, for 2 antenna ports, if transform precoder is enabled and ul-FullPowerTransmission is not configured or configured to fullpowerMode2 or configured to fullpower, or if transform precoder is disabled, maxRank = 1, and and ul-FullPowerTransmission is not configured or configured to fullpowerMode2 or configured to fullpower
//  Table 7.3.1.1.2-5A: Precoding information and number of layers, for 2 antenna ports or Second Precoding information, if transform precoder is enabled and ul-FullPowerTransmission = fullpowerMode1, or if transform precoder is disabled, maxRank = 1, and ul-FullPowerTransmission = fullpowerMode1
var Dci01TpmiAp2Tp1OrTp0MaxRank1 = map[string][]int{
	"0_0": {1, 0},
	"0_1": {1, 1},
	"0_2": {1, 2},
	"0_3": {1, 3},
	"0_4": {1, 4},
	"0_5": {1, 5},
	"0_6": nil,
	"0_7": nil,
	"2_0": {1, 0},
	"2_1": {1, 1},
}

// refer to 3GPP 38.212 vh40
//  Table 7.3.1.1.2-28: SRI indication or Second SRI indication, for non-codebook based PUSCH transmission, Lmax=1
//  Table 7.3.1.1.2-29: SRI indication for non-codebook based PUSCH transmission, Lmax=2
//  Table 7.3.1.1.2-29A: Second SRI indication for non-codebook based PUSCH transmission, Lmax=2
//  Table 7.3.1.1.2-30: SRI indication for non-codebook based PUSCH transmission, Lmax=3
//  Table 7.3.1.1.2-30A: Second SRI indication for non-codebook based PUSCH transmission, Lmax=3
//  Table 7.3.1.1.2-31: SRI indication for non-codebook based PUSCH transmission, Lmax=4
//  Table 7.3.1.1.2-31A: Second SRI indication for non-codebook based PUSCH transmission, Lmax=4
//  key=(Lmax, N_SRS, SRI)
//  2023/2/23: Second SRI indication is not supported!
var Dci01NonCbSri = map[string][]int{
	// Lmax=1
	"1_2_0": {0},
	"1_2_1": {1},
	"1_3_0": {0},
	"1_3_1": {1},
	"1_3_2": {2},
	"1_3_3": nil,
	"1_4_0": {0},
	"1_4_1": {1},
	"1_4_2": {2},
	"1_4_3": {3},
	// Lmax=2
	"2_2_0":  {0},
	"2_2_1":  {1},
	"2_2_2":  {0, 1},
	"2_2_3":  nil,
	"2_3_0":  {0},
	"2_3_1":  {1},
	"2_3_2":  {2},
	"2_3_3":  {0, 1},
	"2_3_4":  {0, 2},
	"2_3_5":  {1, 2},
	"2_3_6":  nil,
	"2_3_7":  nil,
	"2_4_0":  {0},
	"2_4_1":  {1},
	"2_4_2":  {2},
	"2_4_3":  {3},
	"2_4_4":  {0, 1},
	"2_4_5":  {0, 2},
	"2_4_6":  {0, 3},
	"2_4_7":  {1, 2},
	"2_4_8":  {1, 3},
	"2_4_9":  {2, 3},
	"2_4_10": nil,
	"2_4_11": nil,
	"2_4_12": nil,
	"2_4_13": nil,
	"2_4_14": nil,
	"2_4_15": nil,
	// Lmax=3
	"3_2_0":  {0},
	"3_2_1":  {1},
	"3_2_2":  {0, 1},
	"3_2_3":  nil,
	"3_3_0":  {0},
	"3_3_1":  {1},
	"3_3_2":  {2},
	"3_3_3":  {0, 1},
	"3_3_4":  {0, 2},
	"3_3_5":  {1, 2},
	"3_3_6":  {0, 1, 2},
	"3_3_7":  nil,
	"3_4_0":  {0},
	"3_4_1":  {1},
	"3_4_2":  {2},
	"3_4_3":  {3},
	"3_4_4":  {0, 1},
	"3_4_5":  {0, 2},
	"3_4_6":  {0, 3},
	"3_4_7":  {1, 2},
	"3_4_8":  {1, 3},
	"3_4_9":  {2, 3},
	"3_4_10": {0, 1, 2},
	"3_4_11": {0, 1, 3},
	"3_4_12": {0, 2, 3},
	"3_4_13": {1, 2, 3},
	"3_4_14": nil,
	"3_4_15": nil,
	// Lmax=4
	"4_2_0":  {0},
	"4_2_1":  {1},
	"4_2_2":  {0, 1},
	"4_2_3":  nil,
	"4_3_0":  {0},
	"4_3_1":  {1},
	"4_3_2":  {2},
	"4_3_3":  {0, 1},
	"4_3_4":  {0, 2},
	"4_3_5":  {1, 2},
	"4_3_6":  {0, 1, 2},
	"4_3_7":  nil,
	"4_4_0":  {0},
	"4_4_1":  {1},
	"4_4_2":  {2},
	"4_4_3":  {3},
	"4_4_4":  {0, 1},
	"4_4_5":  {0, 2},
	"4_4_6":  {0, 3},
	"4_4_7":  {1, 2},
	"4_4_8":  {1, 3},
	"4_4_9":  {2, 3},
	"4_4_10": {0, 1, 2},
	"4_4_11": {0, 1, 3},
	"4_4_12": {0, 2, 3},
	"4_4_13": {1, 2, 3},
	"4_4_14": {0, 1, 2, 3},
	"4_4_15": nil,
}

// refer to 3GPP 38.331 vf30
// ssb-perRACH-OccasionAndCB-PreamblesPerSSB of RACH-ConfigCommon
var SsbPerRachOccasion2Float = map[string]float64{
	"oneEighth": 0.125,
	"oneFourth": 0.25,
	"oneHalf":   0.5,
	"one":       1,
	"two":       2,
	"four":      4,
	"eight":     8,
	"sixteen":   16,
}

// refer to 3GPP 38.331 vf30
// ssb-perRACH-OccasionAndCB-PreamblesPerSSB of RACH-ConfigCommon
var SsbPerRachOccasion2CbPreamblesPerSsb = map[string][]int{
	"oneEighth": {4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64},
	"oneFourth": {4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64},
	"oneHalf":   {4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64},
	"one":       {4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64},
	"two":       {4, 8, 12, 16, 20, 24, 28, 32},
	"four":      {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	"eight":     {1, 2, 3, 4, 5, 6, 7, 8},
	"sixteen":   {1, 2, 3, 4},
}

// CommonPucchResInfo contains information on common PUCCH resource.
type CommonPucchResInfo struct {
	PucchFmt     int
	FirstSymb    int
	NumSymbs     int
	PrbOffset    int
	InitialCsSet []int
}

// refer to 3GPP 38.213 vf30
//  Table 9.2.1-1: PUCCH resource sets before dedicated PUCCH resource configuration
var CommonPucchResSets = map[int]*CommonPucchResInfo{
	0:  {0, 12, 2, 0, []int{0, 3}},
	1:  {0, 12, 2, 0, []int{0, 4, 8}},
	2:  {0, 12, 2, 3, []int{0, 4, 8}},
	3:  {1, 10, 4, 0, []int{0, 6}},
	4:  {1, 10, 4, 0, []int{0, 3, 6, 9}},
	5:  {1, 10, 4, 2, []int{0, 3, 6, 9}},
	6:  {1, 10, 4, 4, []int{0, 3, 6, 9}},
	7:  {1, 4, 10, 0, []int{0, 6}},
	8:  {1, 4, 10, 0, []int{0, 3, 6, 9}},
	9:  {1, 4, 10, 2, []int{0, 3, 6, 9}},
	10: {1, 4, 10, 4, []int{0, 3, 6, 9}},
	11: {1, 0, 14, 0, []int{0, 6}},
	12: {1, 0, 14, 0, []int{0, 3, 6, 9}},
	13: {1, 0, 14, 2, []int{0, 3, 6, 9}},
	14: {1, 0, 14, 4, []int{0, 3, 6, 9}},
	// Note: for pucch resource index 15, "PRB offset" is floor{N_BWP_size/4}
	15: {1, 0, 14, -1, []int{0, 3, 6, 9}},
}

// CsiRsLocInfo contains information on CSI-RS locations within a slot.
type CsiRsLocInfo struct {
	Row        int
	KBarLBar   [][]int
	Ki         []int
	Li         []int
	CdmGrpIndj []int
	Kap        []int
	Lap        []int
}

// refer to 3GPP 38.211 vf40
//  Table 7.4.1.5.3-1: CSI-RS locations within a slot.
// key=[ports, pdensity, cdm-type], val=list of [row, k-/l-(delta), ki, li, j, k', l']
var CsiRsLoc = map[string][]CsiRsLocInfo{
	"1_3_noCDM":     {{1, [][]int{[]int{0, 0}, []int{4, 0}, []int{8, 0}}, []int{0, 0, 0}, []int{0, 0, 0}, []int{0, 0, 0}, []int{0}, []int{0}}},
	"1_1_noCDM":     {{2, [][]int{[]int{0, 0}}, []int{0}, []int{0}, []int{0}, []int{0}, []int{0}}},
	"1_0.5_noCDM":   {{2, [][]int{[]int{0, 0}}, []int{0}, []int{0}, []int{0}, []int{0}, []int{0}}},
	"2_1_fd-CDM2":   {{3, [][]int{[]int{0, 0}}, []int{0}, []int{0}, []int{0}, []int{0, 1}, []int{0}}},
	"2_0.5_fd-CDM2": {{3, [][]int{[]int{0, 0}}, []int{0}, []int{0}, []int{0}, []int{0, 1}, []int{0}}},
	// "4_1_fd-CDM2": {{4, [][]int{[]int{0, 0}, []int{2, 0}}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0}}},
	// "4_1_fd-CDM2": {{5, [][]int{[]int{0, 0}, []int{0, 1}}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0}}},
	"4_1_fd-CDM2": {{4, [][]int{[]int{0, 0}, []int{2, 0}}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0}}, {5, [][]int{[]int{0, 0}, []int{0, 1}}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0}}},
	// "8_1_fd-CDM2": {{6, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 3}, []int{0, 0, 0, 0}, []int{0, 1, 2, 3}, []int{0, 1}, []int{0}}},
	// "8_1_fd-CDM2": {{7, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}}, []int{0, 1, 0, 1}, []int{0, 0, 0, 0}, []int{0, 1, 2, 3}, []int{0, 1}, []int{0}}},
	"8_1_fd-CDM2":         {{6, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 3}, []int{0, 0, 0, 0}, []int{0, 1, 2, 3}, []int{0, 1}, []int{0}}, {7, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}}, []int{0, 1, 0, 1}, []int{0, 0, 0, 0}, []int{0, 1, 2, 3}, []int{0, 1}, []int{0}}},
	"8_1_cdm4-FD2-TD2":    {{8, [][]int{[]int{0, 0}, []int{0, 0}}, []int{0, 1}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}}},
	"12_1_fd-CDM2":        {{9, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 3, 4, 5}, []int{0, 0, 0, 0, 0, 0}, []int{0, 1, 2, 3, 4, 5}, []int{0, 1}, []int{0}}},
	"12_1_cdm4-FD2-TD2":   {{10, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2}, []int{0, 0, 0}, []int{0, 1, 2}, []int{0, 1}, []int{0, 1}}},
	"16_1_fd-CDM2":        {{11, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}, []int{0, 1}}, []int{0, 1, 2, 3, 0, 1, 2, 3}, []int{0, 0, 0, 0, 0, 0, 0, 0}, []int{0, 1, 2, 3, 4, 5, 6, 7}, []int{0, 1}, []int{0}}},
	"16_0.5_fd-CDM2":      {{11, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}, []int{0, 1}}, []int{0, 1, 2, 3, 0, 1, 2, 3}, []int{0, 0, 0, 0, 0, 0, 0, 0}, []int{0, 1, 2, 3, 4, 5, 6, 7}, []int{0, 1}, []int{0}}},
	"16_1_cdm4-FD2-TD2":   {{12, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 3}, []int{0, 0, 0, 0}, []int{0, 1, 2, 3}, []int{0, 1}, []int{0, 1}}},
	"16_0.5_cdm4-FD2-TD2": {{12, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 3}, []int{0, 0, 0, 0}, []int{0, 1, 2, 3}, []int{0, 1}, []int{0, 1}}},
	"24_1_fd-CDM2":        {{13, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}}, []int{0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2}, []int{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, []int{0, 1}, []int{0}}},
	"24_0.5_fd-CDM2":      {{13, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}}, []int{0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2}, []int{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, []int{0, 1}, []int{0}}},
	"24_1_cdm4-FD2-TD2":   {{14, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 0, 1, 2}, []int{0, 0, 0, 1, 1, 1}, []int{0, 1, 2, 3, 4, 5}, []int{0, 1}, []int{0, 1}}},
	"24_0.5_cdm4-FD2-TD2": {{14, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 0, 1, 2}, []int{0, 0, 0, 1, 1, 1}, []int{0, 1, 2, 3, 4, 5}, []int{0, 1}, []int{0, 1}}},
	"24_1_cdm8-FD2-TD4":   {{15, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2}, []int{0, 0, 0}, []int{0, 1, 2}, []int{0, 1}, []int{0, 1, 2, 3}}},
	"24_0.5_cdm8-FD2-TD4": {{15, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2}, []int{0, 0, 0}, []int{0, 1, 2}, []int{0, 1}, []int{0, 1, 2, 3}}},
	"32_1_fd-CDM2":        {{16, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}, []int{0, 1}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}, []int{0, 1}}, []int{0, 1, 2, 3, 0, 1, 2, 3, 0, 1, 2, 3, 0, 1, 2, 3}, []int{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, []int{0, 1}, []int{0}}},
	"32_0.5_fd-CDM2":      {{16, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}, []int{0, 1}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 1}, []int{0, 1}, []int{0, 1}, []int{0, 1}}, []int{0, 1, 2, 3, 0, 1, 2, 3, 0, 1, 2, 3, 0, 1, 2, 3}, []int{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, []int{0, 1}, []int{0}}},
	"32_1_cdm4-FD2-TD2":   {{17, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 3, 0, 1, 2, 3}, []int{0, 0, 0, 0, 1, 1, 1, 1}, []int{0, 1, 2, 3, 4, 5, 6, 7}, []int{0, 1}, []int{0, 1}}},
	"32_0.5_cdm4-FD2-TD2": {{17, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 3, 0, 1, 2, 3}, []int{0, 0, 0, 0, 1, 1, 1, 1}, []int{0, 1, 2, 3, 4, 5, 6, 7}, []int{0, 1}, []int{0, 1}}},
	"32_1_cdm8-FD2-TD4":   {{18, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 3}, []int{0, 0, 0, 0}, []int{0, 1, 2, 3}, []int{0, 1}, []int{0, 1, 2, 3}}},
	"32_0.5_cdm8-FD2-TD4": {{18, [][]int{[]int{0, 0}, []int{0, 0}, []int{0, 0}, []int{0, 0}}, []int{0, 1, 2, 3}, []int{0, 0, 0, 0}, []int{0, 1, 2, 3}, []int{0, 1}, []int{0, 1, 2, 3}}},
}

// SrsBwInfo contains information on SRS bandwidth configuration.
type SrsBwInfo struct {
	MSRSb []int
	Nb    []int
}

// refer to 3GPP 38.211 vf40
//  Table 6.4.1.4.3-1: SRS bandwidth configuration.
// key=cSrs, val=[mSRSb, Nb]
var SrsBwCfg = map[string]*SrsBwInfo{
	"0":  {[]int{4, 4, 4, 4}, []int{1, 1, 1, 1}},
	"1":  {[]int{8, 4, 4, 4}, []int{1, 2, 1, 1}},
	"2":  {[]int{12, 4, 4, 4}, []int{1, 3, 1, 1}},
	"3":  {[]int{16, 4, 4, 4}, []int{1, 4, 1, 1}},
	"4":  {[]int{16, 8, 4, 4}, []int{1, 2, 2, 1}},
	"5":  {[]int{20, 4, 4, 4}, []int{1, 5, 1, 1}},
	"6":  {[]int{24, 4, 4, 4}, []int{1, 6, 1, 1}},
	"7":  {[]int{24, 12, 4, 4}, []int{1, 2, 3, 1}},
	"8":  {[]int{28, 4, 4, 4}, []int{1, 7, 1, 1}},
	"9":  {[]int{32, 16, 8, 4}, []int{1, 2, 2, 2}},
	"10": {[]int{36, 12, 4, 4}, []int{1, 3, 3, 1}},
	"11": {[]int{40, 20, 4, 4}, []int{1, 2, 5, 1}},
	"12": {[]int{48, 16, 8, 4}, []int{1, 3, 2, 2}},
	"13": {[]int{48, 24, 12, 4}, []int{1, 2, 2, 3}},
	"14": {[]int{52, 4, 4, 4}, []int{1, 13, 1, 1}},
	"15": {[]int{56, 28, 4, 4}, []int{1, 2, 7, 1}},
	"16": {[]int{60, 20, 4, 4}, []int{1, 3, 5, 1}},
	"17": {[]int{64, 32, 16, 4}, []int{1, 2, 2, 4}},
	"18": {[]int{72, 24, 12, 4}, []int{1, 3, 2, 3}},
	"19": {[]int{72, 36, 12, 4}, []int{1, 2, 3, 3}},
	"20": {[]int{76, 4, 4, 4}, []int{1, 19, 1, 1}},
	"21": {[]int{80, 40, 20, 4}, []int{1, 2, 2, 5}},
	"22": {[]int{88, 44, 4, 4}, []int{1, 2, 11, 1}},
	"23": {[]int{96, 32, 16, 4}, []int{1, 3, 2, 4}},
	"24": {[]int{96, 48, 24, 4}, []int{1, 2, 2, 6}},
	"25": {[]int{104, 52, 4, 4}, []int{1, 2, 13, 1}},
	"26": {[]int{112, 56, 28, 4}, []int{1, 2, 2, 7}},
	"27": {[]int{120, 60, 20, 4}, []int{1, 2, 3, 5}},
	"28": {[]int{120, 40, 8, 4}, []int{1, 3, 5, 2}},
	"29": {[]int{120, 24, 12, 4}, []int{1, 5, 2, 3}},
	"30": {[]int{128, 64, 32, 4}, []int{1, 2, 2, 8}},
	"31": {[]int{128, 64, 16, 4}, []int{1, 2, 4, 4}},
	"32": {[]int{128, 16, 8, 4}, []int{1, 8, 2, 2}},
	"33": {[]int{132, 44, 4, 4}, []int{1, 3, 11, 1}},
	"34": {[]int{136, 68, 4, 4}, []int{1, 2, 17, 1}},
	"35": {[]int{144, 72, 36, 4}, []int{1, 2, 2, 9}},
	"36": {[]int{144, 48, 24, 12}, []int{1, 3, 2, 2}},
	"37": {[]int{144, 48, 16, 4}, []int{1, 3, 3, 4}},
	"38": {[]int{144, 16, 8, 4}, []int{1, 9, 2, 2}},
	"39": {[]int{152, 76, 4, 4}, []int{1, 2, 19, 1}},
	"40": {[]int{160, 80, 40, 4}, []int{1, 2, 2, 10}},
	"41": {[]int{160, 80, 20, 4}, []int{1, 2, 4, 5}},
	"42": {[]int{160, 32, 16, 4}, []int{1, 5, 2, 4}},
	"43": {[]int{168, 84, 28, 4}, []int{1, 2, 3, 7}},
	"44": {[]int{176, 88, 44, 4}, []int{1, 2, 2, 11}},
	"45": {[]int{184, 92, 4, 4}, []int{1, 2, 23, 1}},
	"46": {[]int{192, 96, 48, 4}, []int{1, 2, 2, 12}},
	"47": {[]int{192, 96, 24, 4}, []int{1, 2, 4, 6}},
	"48": {[]int{192, 64, 16, 4}, []int{1, 3, 4, 4}},
	"49": {[]int{192, 24, 8, 4}, []int{1, 8, 3, 2}},
	"50": {[]int{208, 104, 52, 4}, []int{1, 2, 2, 13}},
	"51": {[]int{216, 108, 36, 4}, []int{1, 2, 3, 9}},
	"52": {[]int{224, 112, 56, 4}, []int{1, 2, 2, 14}},
	"53": {[]int{240, 120, 60, 4}, []int{1, 2, 2, 15}},
	"54": {[]int{240, 80, 20, 4}, []int{1, 3, 4, 5}},
	"55": {[]int{240, 48, 16, 8}, []int{1, 5, 3, 2}},
	"56": {[]int{240, 24, 12, 4}, []int{1, 10, 2, 3}},
	"57": {[]int{256, 128, 64, 4}, []int{1, 2, 2, 16}},
	"58": {[]int{256, 128, 32, 4}, []int{1, 2, 4, 8}},
	"59": {[]int{256, 16, 8, 4}, []int{1, 16, 2, 2}},
	"60": {[]int{264, 132, 44, 4}, []int{1, 2, 3, 11}},
	"61": {[]int{272, 136, 68, 4}, []int{1, 2, 2, 17}},
	"62": {[]int{272, 68, 4, 4}, []int{1, 4, 17, 1}},
	"63": {[]int{272, 16, 8, 4}, []int{1, 17, 2, 2}},
}

// offset of CORESET0 w.r.t. SSB
var Coreset0Offset = 0

// minimum channel bandwidth
var MinChBw = 0

// starting PRB of RMSI/CORESET0 w.r.t. first usable PRB of carrier
var Coreset0StartRb = 0

// valid PUSCH PRB allocations when transforming precoding is enabled
var LrbsMsg3PuschTp = []int{}
var LrbsDedPuschTp = []int{}

// SLIV look-up tables for PDSCH
var PdschToSliv, PdschFromSliv = initPdschSliv()

// SLIV look-up tables for PUSCH
var PuschToSliv, PuschFromSliv = initPuschSliv()

// constants
const (
	NumScPerPrb = 12
)

func initPdschSliv() (map[string]int, map[string][]int) {
	// prefix
	// "00": mapping type A + normal cp
	// "01": mapping type A + extended cp
	// "10": mapping type B + normal cp
	// "11": mapping type B + extended cp
	pdschToSliv := make(map[string]int)
	pdschFromSliv := make(map[string][]int)
	var prefix string

	// case #1: prefix="00"
	prefix = "00"
	for _, S := range utils.PyRange(0, 4, 1) {
		for _, L := range utils.PyRange(3, 15, 1) {
			if S+L >= 3 && S+L < 15 {
				sliv, err := makeSliv(S, L)
				if err == nil {
					keyToSliv := fmt.Sprintf("%s_%d_%d", prefix, S, L)
					pdschToSliv[keyToSliv] = sliv
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					pdschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	// case #2: prefix="01"
	prefix = "01"
	for _, S := range utils.PyRange(0, 4, 1) {
		for _, L := range utils.PyRange(3, 13, 1) {
			if S+L >= 3 && S+L < 13 {
				sliv, err := makeSliv(S, L)
				if err == nil {
					keyToSliv := fmt.Sprintf("%s_%d_%d", prefix, S, L)
					pdschToSliv[keyToSliv] = sliv
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					pdschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	// case #3: prefix="10"
	prefix = "10"
	for _, S := range utils.PyRange(0, 13, 1) {
		for _, L := range []int{2, 4, 7} {
			if S+L >= 2 && S+L < 15 {
				sliv, err := makeSliv(S, L)
				if err == nil {
					keyToSliv := fmt.Sprintf("%s_%d_%d", prefix, S, L)
					pdschToSliv[keyToSliv] = sliv
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					pdschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	// case #4: prefix="11"
	prefix = "11"
	for _, S := range utils.PyRange(0, 11, 1) {
		for _, L := range []int{2, 4, 6} {
			if S+L >= 2 && S+L < 13 {
				sliv, err := makeSliv(S, L)
				if err == nil {
					keyToSliv := fmt.Sprintf("%s_%d_%d", prefix, S, L)
					pdschToSliv[keyToSliv] = sliv
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					pdschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	return pdschToSliv, pdschFromSliv
}

func initPuschSliv() (map[string]int, map[string][]int) {
	// prefix
	// "00": mapping type A + normal cp
	// "01": mapping type A + extended cp
	// "10": mapping type B + normal cp
	// "11": mapping type B + extended cp
	puschToSliv := make(map[string]int)
	puschFromSliv := make(map[string][]int)
	var prefix string

	// case #1: prefix="00"
	prefix = "00"
	for _, S := range []int{0} {
		for _, L := range utils.PyRange(4, 15, 1) {
			if S+L >= 4 && S+L < 15 {
				sliv, err := makeSliv(S, L)
				if err == nil {
					keyToSliv := fmt.Sprintf("%s_%d_%d", prefix, S, L)
					puschToSliv[keyToSliv] = sliv
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					puschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	// case #2: prefix="01"
	prefix = "01"
	for _, S := range []int{0} {
		for _, L := range utils.PyRange(4, 13, 1) {
			if S+L >= 4 && S+L < 13 {
				sliv, err := makeSliv(S, L)
				if err == nil {
					keyToSliv := fmt.Sprintf("%s_%d_%d", prefix, S, L)
					puschToSliv[keyToSliv] = sliv
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					puschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	// case #3: prefix="10"
	prefix = "10"
	for _, S := range utils.PyRange(0, 14, 1) {
		for _, L := range utils.PyRange(1, 15, 1) {
			if S+L >= 1 && S+L < 15 {
				sliv, err := makeSliv(S, L)
				if err == nil {
					keyToSliv := fmt.Sprintf("%s_%d_%d", prefix, S, L)
					puschToSliv[keyToSliv] = sliv
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					puschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	// case #4: prefix="11"
	prefix = "11"
	for _, S := range utils.PyRange(0, 13, 1) {
		for _, L := range utils.PyRange(1, 13, 1) {
			if S+L >= 1 && S+L < 13 {
				sliv, err := makeSliv(S, L)
				if err == nil {
					keyToSliv := fmt.Sprintf("%s_%d_%d", prefix, S, L)
					puschToSliv[keyToSliv] = sliv
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					puschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	return puschToSliv, puschFromSliv
}

func makeSliv(S, L int) (int, error) {
	if L <= 0 || L > 14-S {
		return -1, errors.New(fmt.Sprintf("invalid S/L combination: S=%d, L=%d", S, L))
	}

	sliv := 0
	if (L - 1) <= 7 {
		sliv = 14*(L-1) + S
	} else {
		sliv = 14*(14-L+1) + (14-1-S)
	}

	return sliv, nil
}


// ToSliv returns SLIV from given S/L.
//  S: starting symbol
//	L: number of symbols
//	sch: PDSCH or PUSCH
//	mappingType: typeA or typeB
//	cp: normal or extended
func ToSliv(S, L int, sch, mappingType, cp string) (int, error) {
	var prefix string
	if mappingType == "typeA" {
		prefix += "0"
	} else {
	    prefix += "1"
	}

	if cp == "normal" {
		prefix += "0"
	} else {
		prefix += "1"
	}

	key := fmt.Sprintf("%v_%v_%v", prefix, S, L)
	var sliv int
	var exist bool
	if sch == "PDSCH" {
	    sliv, exist = PdschToSliv[key]
	} else {
		sliv, exist = PuschToSliv[key]
	}

	if !exist {
		return -1, errors.New(fmt.Sprintf("Invalid call to ToSliv: prefix=%v, S=%v, L=%v", prefix, S, L))
	} else {
		return sliv, nil
	}
}

// FromSliv returns S/L from given sliv.
//  sliv: start and length indicator
//	sch: PDSCH or PUSCH
//	mappingType: typeA or typeB
//	cp: normal or extended
func FromSliv(sliv int, sch, mappingType, cp string) ([]int, error) {
	var prefix string
	if mappingType == "typeA" {
		prefix += "0"
	} else {
		prefix += "1"
	}

	if cp == "normal" {
		prefix += "0"
	} else {
		prefix += "1"
	}

	key := fmt.Sprintf("%v_%v", prefix, sliv)
	var sl []int
	var exist bool
	if sch == "PDSCH" {
		sl, exist = PdschFromSliv[key]
	} else {
		sl, exist = PuschFromSliv[key]
	}

	if !exist {
		return nil, errors.New(fmt.Sprintf("Invalid call to FromSliv: prefix=%v, slIV=%v", prefix, sliv))
	} else {
		return sl, nil
	}
}

/*
# initialize SLIV look-up tables
        self.initPdschSliv()
        self.initPuschSliv()
        '''
        self.ngwin.logEdit.append('contents of self.nrPdschToSliv: ')
        for key, val in self.nrPdschToSliv.items():
            prefix, S, L = key.split('_')
            self.ngwin.logEdit.append('%s, %s, %s, %s'%(prefix, S, L, val))
        self.ngwin.logEdit.append('contents of self.nrPuschToSliv: ')
        for key, val in self.nrPuschToSliv.items():
            prefix, S, L = key.split('_')
            self.ngwin.logEdit.append('%s, %s, %s, %s'%(prefix, S, L, val))
        '''

*/
