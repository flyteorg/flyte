import { addons } from '@storybook/manager-api'
import { create } from '@storybook/theming'

const logoSvg = `data:image/svg+xml,${encodeURIComponent(`
<svg xmlns="http://www.w3.org/2000/svg" width="300" height="65" viewBox="0 0 300 65" style="fill: #FCB51D;">
<title>Union.ai</title>
<path style="fill: #FCB51D;" d="M32,64.8C14.4,64.8,0,51.5,0,34V3.6h17.6v41.3c0,1.9,1.1,3,3,3h23c1.9,0,3-1.1,3-3V3.6H64V34
	C64,51.5,49.6,64.8,32,64.8z M69.9,30.9v30.4h17.6V20c0-1.9,1.1-3,3-3h23c1.9,0,3,1.1,3,3v41.3H134V30.9c0-17.5-14.4-30.8-32.1-30.8
	S69.9,13.5,69.9,30.9z M236,30.9v30.4h17.6V20c0-1.9,1.1-3,3-3h23c1.9,0,3,1.1,3,3v41.3H300V30.9c0-17.5-14.4-30.8-32-30.8
	S236,13.5,236,30.9L236,30.9z M230.1,32.4c0,18.2-14.2,32.5-32.2,32.5s-32-14.3-32-32.5s14-32.1,32-32.1S230.1,14.3,230.1,32.4
	L230.1,32.4z M213.5,20.2c0-1.9-1.1-3-3-3h-24.8c-1.9,0-3,1.1-3,3v24.5c0,1.9,1.1,3,3,3h24.8c1.9,0,3-1.1,3-3V20.2z M158.9,3.6
	h-17.6v57.8h17.6V3.6z"></path>
</svg>
`)}`

addons.setConfig({
  theme: create({
    base: 'dark',
    brandTitle: 'Union AI',
    brandImage: logoSvg,
    brandTarget: '_self',
    colorPrimary: '#FCB51F',
    colorSecondary: '#FCB51F',

    // UI
    appBg: '#1a1a1a',
    appContentBg: '#1a1a1a',
    appBorderColor: 'rgba(255,255,255,.1)',
    appBorderRadius: 4,

    // Text colors
    textColor: '#FFFFFF',
    textInverseColor: '#1A1A1A',

    // Toolbar default and active colors
    barTextColor: '#999999',
    barSelectedColor: '#FCB51F',
    barBg: '#1a1a1a',

    // Form colors
    inputBg: '#2A2A2A',
    inputBorder: 'rgba(255,255,255,.1)',
    inputTextColor: '#FFFFFF',
    inputBorderRadius: 4,
  }),
})
