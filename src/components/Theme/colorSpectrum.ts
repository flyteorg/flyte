export type ColorSpectrumType =
    | 'white'
    | 'black'
    | 'red0'
    | 'red5'
    | 'red10'
    | 'red20'
    | 'red30'
    | 'red40'
    | 'red50'
    | 'red60'
    | 'red70'
    | 'red80'
    | 'red90'
    | 'red100'
    | 'sunset0'
    | 'sunset5'
    | 'sunset10'
    | 'sunset20'
    | 'sunset30'
    | 'sunset40'
    | 'sunset50'
    | 'sunset60'
    | 'sunset70'
    | 'sunset80'
    | 'sunset90'
    | 'sunset100'
    | 'orange0'
    | 'orange5'
    | 'orange10'
    | 'orange20'
    | 'orange30'
    | 'orange40'
    | 'orange50'
    | 'orange60'
    | 'orange70'
    | 'orange80'
    | 'orange90'
    | 'orange100'
    | 'amber0'
    | 'amber5'
    | 'amber10'
    | 'amber20'
    | 'amber30'
    | 'amber40'
    | 'amber50'
    | 'amber60'
    | 'amber70'
    | 'amber80'
    | 'amber90'
    | 'amber100'
    | 'yellow0'
    | 'yellow5'
    | 'yellow10'
    | 'yellow20'
    | 'yellow30'
    | 'yellow40'
    | 'yellow50'
    | 'yellow60'
    | 'yellow70'
    | 'yellow80'
    | 'yellow90'
    | 'yellow100'
    | 'citron0'
    | 'citron5'
    | 'citron10'
    | 'citron20'
    | 'citron30'
    | 'citron40'
    | 'citron50'
    | 'citron60'
    | 'citron70'
    | 'citron80'
    | 'citron90'
    | 'citron100'
    | 'lime0'
    | 'lime5'
    | 'lime10'
    | 'lime20'
    | 'lime30'
    | 'lime40'
    | 'lime50'
    | 'lime60'
    | 'lime70'
    | 'lime80'
    | 'lime90'
    | 'lime100'
    | 'green0'
    | 'green5'
    | 'green10'
    | 'green20'
    | 'green30'
    | 'green40'
    | 'green50'
    | 'green60'
    | 'green70'
    | 'green80'
    | 'green90'
    | 'green100'
    | 'mint0'
    | 'mint5'
    | 'mint10'
    | 'mint20'
    | 'mint30'
    | 'mint40'
    | 'mint50'
    | 'mint60'
    | 'mint70'
    | 'mint80'
    | 'mint90'
    | 'mint100'
    | 'teal0'
    | 'teal5'
    | 'teal10'
    | 'teal20'
    | 'teal30'
    | 'teal40'
    | 'teal50'
    | 'teal60'
    | 'teal70'
    | 'teal80'
    | 'teal90'
    | 'teal100'
    | 'cyan0'
    | 'cyan5'
    | 'cyan10'
    | 'cyan20'
    | 'cyan30'
    | 'cyan40'
    | 'cyan50'
    | 'cyan60'
    | 'cyan70'
    | 'cyan80'
    | 'cyan90'
    | 'cyan100'
    | 'blue0'
    | 'blue5'
    | 'blue10'
    | 'blue20'
    | 'blue30'
    | 'blue40'
    | 'blue50'
    | 'blue60'
    | 'blue70'
    | 'blue80'
    | 'blue90'
    | 'blue100'
    | 'indigo0'
    | 'indigo5'
    | 'indigo10'
    | 'indigo20'
    | 'indigo30'
    | 'indigo40'
    | 'indigo50'
    | 'indigo60'
    | 'indigo70'
    | 'indigo80'
    | 'indigo90'
    | 'indigo100'
    | 'purple0'
    | 'purple5'
    | 'purple10'
    | 'purple20'
    | 'purple30'
    | 'purple40'
    | 'purple50'
    | 'purple60'
    | 'purple70'
    | 'purple80'
    | 'purple90'
    | 'purple100'
    | 'pink0'
    | 'pink5'
    | 'pink10'
    | 'pink20'
    | 'pink30'
    | 'pink40'
    | 'pink50'
    | 'pink60'
    | 'pink70'
    | 'pink80'
    | 'pink90'
    | 'pink100'
    | 'rose0'
    | 'rose5'
    | 'rose10'
    | 'rose20'
    | 'rose30'
    | 'rose40'
    | 'rose50'
    | 'rose60'
    | 'rose70'
    | 'rose80'
    | 'rose90'
    | 'rose100'
    | 'gray0'
    | 'gray5'
    | 'gray10'
    | 'gray15'
    | 'gray20'
    | 'gray25'
    | 'gray30'
    | 'gray40'
    | 'gray50'
    | 'gray60'
    | 'gray70'
    | 'gray80'
    | 'gray90'
    | 'gray100';

export const COLOR_SPECTRUM: {
    [key in ColorSpectrumType]: { color: string; value: number };
} = {
    white: {
        color: '#FFFFFF',
        value: 0
    },
    black: {
        color: '#000000',
        value: 100
    },
    red0: {
        color: '#FFFAFB',
        value: 0
    },
    red5: {
        color: '#FFE5E9',
        value: 5
    },
    red10: {
        color: '#FFCFD5',
        value: 10
    },
    red20: {
        color: '#FFA0AC',
        value: 20
    },
    red30: {
        color: '#FF7689',
        value: 30
    },
    red40: {
        color: '#FF516B',
        value: 40
    },
    red50: {
        color: '#FF3354',
        value: 50
    },
    red60: {
        color: '#E6193F',
        value: 60
    },
    red70: {
        color: '#B8072C',
        value: 70
    },
    red80: {
        color: '#8C0020',
        value: 80
    },
    red90: {
        color: '#670019',
        value: 90
    },
    red100: {
        color: '#560015',
        value: 100
    },
    sunset0: {
        color: '#FFFBFA',
        value: 0
    },
    sunset5: {
        color: '#FFE4DD',
        value: 5
    },
    sunset10: {
        color: '#FFCCBF',
        value: 10
    },
    sunset20: {
        color: '#FF9E87',
        value: 20
    },
    sunset30: {
        color: '#FF7B5C',
        value: 30
    },
    sunset40: {
        color: '#FF623E',
        value: 40
    },
    sunset50: {
        color: '#FF4E28',
        value: 50
    },
    sunset60: {
        color: '#DB3615',
        value: 60
    },
    sunset70: {
        color: '#AF230A',
        value: 70
    },
    sunset80: {
        color: '#841604',
        value: 80
    },
    sunset90: {
        color: '#5F0E01',
        value: 90
    },
    sunset100: {
        color: '#4E0B00',
        value: 100
    },
    orange0: {
        color: '#FFF6F2',
        value: 0
    },
    orange5: {
        color: '#FFE8DD',
        value: 5
    },
    orange10: {
        color: '#FFD9C7',
        value: 10
    },
    orange20: {
        color: '#FFB38F',
        value: 20
    },
    orange30: {
        color: '#FF915D',
        value: 30
    },
    orange40: {
        color: '#FF7232',
        value: 40
    },
    orange50: {
        color: '#F9560E',
        value: 50
    },
    orange60: {
        color: '#D03D00',
        value: 60
    },
    orange70: {
        color: '#A82E00',
        value: 70
    },
    orange80: {
        color: '#832300',
        value: 80
    },
    orange90: {
        color: '#651A00',
        value: 90
    },
    orange100: {
        color: '#581600',
        value: 100
    },
    amber0: {
        color: '#FFFDFA',
        value: 0
    },
    amber5: {
        color: '#FFF6E7',
        value: 5
    },
    amber10: {
        color: '#FFF0D4',
        value: 10
    },
    amber20: {
        color: '#FFE0A9',
        value: 20
    },
    amber30: {
        color: '#FFD082',
        value: 30
    },
    amber40: {
        color: '#FFC161',
        value: 40
    },
    amber50: {
        color: '#FFB146',
        value: 50
    },
    amber60: {
        color: '#FFA030',
        value: 60
    },
    amber70: {
        color: '#FF8D1F',
        value: 70
    },
    amber80: {
        color: '#FE7E13',
        value: 80
    },
    amber90: {
        color: '#E66909',
        value: 90
    },
    amber100: {
        color: '#CB5803',
        value: 100
    },
    yellow0: {
        color: '#FFFEFA',
        value: 0
    },
    yellow5: {
        color: '#FFF8D9',
        value: 5
    },
    yellow10: {
        color: '#FFF3B8',
        value: 10
    },
    yellow20: {
        color: '#FFE77B',
        value: 20
    },
    yellow30: {
        color: '#FFDD4C',
        value: 30
    },
    yellow40: {
        color: '#FFD32A',
        value: 40
    },
    yellow50: {
        color: '#FFCA13',
        value: 50
    },
    yellow60: {
        color: '#FFC002',
        value: 60
    },
    yellow70: {
        color: '#EFAC00',
        value: 70
    },
    yellow80: {
        color: '#DC9900',
        value: 80
    },
    yellow90: {
        color: '#C78700',
        value: 90
    },
    yellow100: {
        color: '#B07600',
        value: 100
    },
    citron0: {
        color: '#FFFFF2',
        value: 0
    },
    citron5: {
        color: '#FFFFD2',
        value: 5
    },
    citron10: {
        color: '#FEFFB2',
        value: 10
    },
    citron20: {
        color: '#FBFF6F',
        value: 20
    },
    citron30: {
        color: '#F1FB3B',
        value: 30
    },
    citron40: {
        color: '#E2F316',
        value: 40
    },
    citron50: {
        color: '#CCE700',
        value: 50
    },
    citron60: {
        color: '#B5D900',
        value: 60
    },
    citron70: {
        color: '#9AC800',
        value: 70
    },
    citron80: {
        color: '#82B400',
        value: 80
    },
    citron90: {
        color: '#6C9C00',
        value: 90
    },
    citron100: {
        color: '#578000',
        value: 100
    },
    lime0: {
        color: '#FDFFFA',
        value: 0
    },
    lime5: {
        color: '#EDFED0',
        value: 5
    },
    lime10: {
        color: '#D6F3A0',
        value: 10
    },
    lime20: {
        color: '#A4DC48',
        value: 20
    },
    lime30: {
        color: '#75C404',
        value: 30
    },
    lime40: {
        color: '#5EAB00',
        value: 40
    },
    lime50: {
        color: '#499300',
        value: 50
    },
    lime60: {
        color: '#347D00',
        value: 60
    },
    lime70: {
        color: '#216800',
        value: 70
    },
    lime80: {
        color: '#155600',
        value: 80
    },
    lime90: {
        color: '#0E4400',
        value: 90
    },
    lime100: {
        color: '#0A3600',
        value: 100
    },
    green0: {
        color: '#FAFFFC',
        value: 0
    },
    green5: {
        color: '#D1FFE2',
        value: 5
    },
    green10: {
        color: '#A8FFC4',
        value: 10
    },
    green20: {
        color: '#4BE77A',
        value: 20
    },
    green30: {
        color: '#04CD3D',
        value: 30
    },
    green40: {
        color: '#00B32E',
        value: 40
    },
    green50: {
        color: '#009B22',
        value: 50
    },
    green60: {
        color: '#008316',
        value: 60
    },
    green70: {
        color: '#006E0B',
        value: 70
    },
    green80: {
        color: '#005A05',
        value: 80
    },
    green90: {
        color: '#004802',
        value: 90
    },
    green100: {
        color: '#003901',
        value: 100
    },
    mint0: {
        color: '#FAFFFD',
        value: 0
    },
    mint5: {
        color: '#D1FFEE',
        value: 5
    },
    mint10: {
        color: '#A6FBDE',
        value: 10
    },
    mint20: {
        color: '#4AE3AE',
        value: 20
    },
    mint30: {
        color: '#04CA83',
        value: 30
    },
    mint40: {
        color: '#00B16F',
        value: 40
    },
    mint50: {
        color: '#00985D',
        value: 50
    },
    mint60: {
        color: '#00824C',
        value: 60
    },
    mint70: {
        color: '#006C3C',
        value: 70
    },
    mint80: {
        color: '#00592F',
        value: 80
    },
    mint90: {
        color: '#004724',
        value: 90
    },
    mint100: {
        color: '#00381C',
        value: 100
    },
    teal0: {
        color: '#FAFFFE',
        value: 0
    },
    teal5: {
        color: '#D1FFF7',
        value: 5
    },
    teal10: {
        color: '#A8FFF4',
        value: 10
    },
    teal20: {
        color: '#4CEAE4',
        value: 20
    },
    teal30: {
        color: '#04CED2',
        value: 30
    },
    teal40: {
        color: '#00B0B9',
        value: 40
    },
    teal50: {
        color: '#00949F',
        value: 50
    },
    teal60: {
        color: '#007B85',
        value: 60
    },
    teal70: {
        color: '#00626B',
        value: 70
    },
    teal80: {
        color: '#004C53',
        value: 80
    },
    teal90: {
        color: '#003B40',
        value: 90
    },
    teal100: {
        color: '#003338',
        value: 100
    },
    cyan0: {
        color: '#FAFDFF',
        value: 0
    },
    cyan5: {
        color: '#E7F6FF',
        value: 5
    },
    cyan10: {
        color: '#D4F0FF',
        value: 10
    },
    cyan20: {
        color: '#A9E1FF',
        value: 20
    },
    cyan30: {
        color: '#82D2FF',
        value: 30
    },
    cyan40: {
        color: '#5DBCF4',
        value: 40
    },
    cyan50: {
        color: '#3A97D3',
        value: 50
    },
    cyan60: {
        color: '#2277B3',
        value: 60
    },
    cyan70: {
        color: '#135B96',
        value: 70
    },
    cyan80: {
        color: '#09457B',
        value: 80
    },
    cyan90: {
        color: '#043563',
        value: 90
    },
    cyan100: {
        color: '#01284E',
        value: 100
    },
    blue0: {
        color: '#FAFBFF',
        value: 0
    },
    blue5: {
        color: '#E8ECFF',
        value: 5
    },
    blue10: {
        color: '#D5DCFF',
        value: 10
    },
    blue20: {
        color: '#ACBBFF',
        value: 20
    },
    blue30: {
        color: '#869DFF',
        value: 30
    },
    blue40: {
        color: '#6686FF',
        value: 40
    },
    blue50: {
        color: '#4B73FF',
        value: 50
    },
    blue60: {
        color: '#3668FF',
        value: 60
    },
    blue70: {
        color: '#2156DB',
        value: 70
    },
    blue80: {
        color: '#1242AF',
        value: 80
    },
    blue90: {
        color: '#093186',
        value: 90
    },
    blue100: {
        color: '#042260',
        value: 100
    },
    indigo0: {
        color: '#FAFAFF',
        value: 0
    },
    indigo5: {
        color: '#EBEBFF',
        value: 5
    },
    indigo10: {
        color: '#DCDCFF',
        value: 10
    },
    indigo20: {
        color: '#BABAFF',
        value: 20
    },
    indigo30: {
        color: '#9C9BFF',
        value: 30
    },
    indigo40: {
        color: '#8481FF',
        value: 40
    },
    indigo50: {
        color: '#726BFF',
        value: 50
    },
    indigo60: {
        color: '#665AFF',
        value: 60
    },
    indigo70: {
        color: '#604CFF',
        value: 70
    },
    indigo80: {
        color: '#523BE4',
        value: 80
    },
    indigo90: {
        color: '#3E29B1',
        value: 90
    },
    indigo100: {
        color: '#2B1B81',
        value: 100
    },
    purple0: {
        color: '#FDFAFF',
        value: 0
    },
    purple5: {
        color: '#F6EBFF',
        value: 5
    },
    purple10: {
        color: '#ECDCFF',
        value: 10
    },
    purple20: {
        color: '#D7B8FF',
        value: 20
    },
    purple30: {
        color: '#C294FF',
        value: 30
    },
    purple40: {
        color: '#AD71FF',
        value: 40
    },
    purple50: {
        color: '#9B52FF',
        value: 50
    },
    purple60: {
        color: '#8B37FF',
        value: 60
    },
    purple70: {
        color: '#7B20F9',
        value: 70
    },
    purple80: {
        color: '#590DC4',
        value: 80
    },
    purple90: {
        color: '#420499',
        value: 90
    },
    purple100: {
        color: '#390188',
        value: 100
    },
    pink0: {
        color: '#FFFAFD',
        value: 0
    },
    pink5: {
        color: '#FFE1F2',
        value: 5
    },
    pink10: {
        color: '#FFC7E4',
        value: 10
    },
    pink20: {
        color: '#FF8FCC',
        value: 20
    },
    pink30: {
        color: '#FF5DBB',
        value: 30
    },
    pink40: {
        color: '#FF32B1',
        value: 40
    },
    pink50: {
        color: '#FF0EB0',
        value: 50
    },
    pink60: {
        color: '#DE00A7',
        value: 60
    },
    pink70: {
        color: '#BD00A0',
        value: 70
    },
    pink80: {
        color: '#A00093',
        value: 80
    },
    pink90: {
        color: '#860081',
        value: 90
    },
    pink100: {
        color: '#71006F',
        value: 100
    },
    rose0: {
        color: '#FFF2F5',
        value: 0
    },
    rose5: {
        color: '#FFE1E9',
        value: 5
    },
    rose10: {
        color: '#FFCFDC',
        value: 10
    },
    rose20: {
        color: '#FFA0BA',
        value: 20
    },
    rose30: {
        color: '#FF769E',
        value: 30
    },
    rose40: {
        color: '#FF5187',
        value: 40
    },
    rose50: {
        color: '#FF3378',
        value: 50
    },
    rose60: {
        color: '#E51966',
        value: 60
    },
    rose70: {
        color: '#B70752',
        value: 70
    },
    rose80: {
        color: '#8B0040',
        value: 80
    },
    rose90: {
        color: '#660031',
        value: 90
    },
    rose100: {
        color: '#55002A',
        value: 100
    },
    gray0: {
        color: '#FCFCFF',
        value: 0
    },
    gray5: {
        color: '#F4F4FA',
        value: 5
    },
    gray10: {
        color: '#E7E7EF',
        value: 10
    },
    gray15: {
        color: '#D8D8E4',
        value: 15
    },
    gray20: {
        color: '#CACAD9',
        value: 20
    },
    gray25: {
        color: '#B3B3B3',
        value: 25
    },
    gray30: {
        color: '#ACACC0',
        value: 30
    },
    gray40: {
        color: '#9191A8',
        value: 40
    },
    gray50: {
        color: '#787891',
        value: 50
    },
    gray60: {
        color: '#63637B',
        value: 60
    },
    gray70: {
        color: '#515167',
        value: 70
    },
    gray80: {
        color: '#414155',
        value: 80
    },
    gray90: {
        color: '#333344',
        value: 90
    },
    gray100: {
        color: '#292936',
        value: 100
    }
};
