package utils

import "html/template"

func IncludeSvg(name string) template.HTML {
	svg := ""
	switch name {
	case "brand_svg":
		svg = `
			<svg style="width: 22px; height: 22px; margin-bottom: .55rem;" id="Ebene_1" data-name="Ebene 1" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120">
				<defs>
				<style>
					.cls-1 {
					fill: #1d1d1b;
					}

					.cls-2 {
					fill: #fff;
					}
					[data-theme="dark"] .cls-1 {
					fill: #fff;
					}

					[data-theme="dark"] .cls-2 {
					fill: #1e1c1f;
					}
				</style>
				</defs>
				<rect class="cls-1" x="-14.88" y="-19.92" width="149.75" height="146.77" />
				<path
				class="cls-2"
				d="M59-12.9c-47.25,0-66.43-4.44-66.43,32.35C-7.45,45-8.17,120-8.17,120H22.56S35.78,94.52,36.41,93.86c1.36-1.42,20.81,9,22,10.23.25.26,5.4,15.91,5.4,15.91h64.48s-.15-68-.15-90.77C128.17-15.07,99.59-12.9,59-12.9ZM84.32,88c-16.59,13.26-43.13,6.9-51-15.71a36.18,36.18,0,0,1-1.21-19.85c.85-4,2.2-4.57,5.49-2.37,7,4.68,13.88,9.24,21.36,14.22.73-1.33,1.7-2.6,2.35-4.05a4.31,4.31,0,0,0,.31-2.3c-.5-3.11,1-5.85,3.83-6.55a5.18,5.18,0,0,1,6.24,4,5.48,5.48,0,0,1-3.93,6.44,2.48,2.48,0,0,0-1.24.65c-.17.26-3.07,4.92-3.07,4.92L73.72,74.5c3.47,2.39,6.91,4.76,10.41,7.1S87.52,85.42,84.32,88ZM66.2,39.23a16.16,16.16,0,0,1,16.27,9.91c.06.15.12.31.19.46.66,1.53.42,2-1.22,2.44l-.59.15c-1.08.28-1.25-.7-2.07-2.57a11.47,11.47,0,0,0-11.87-6.85c-1.65.18-2.79.48-3-.39C63.25,39.65,63.37,39.49,66.2,39.23ZM89.32,50l-.92.24c-1.68.44-2-1.13-3.25-4.13A18,18,0,0,0,66.62,35c-2.55.26-4.32.72-4.66-.67-1.05-4.37-.87-4.62,3.53-5a25.51,25.51,0,0,1,25.4,16c.09.25.18.5.29.75C92.22,48.55,91.86,49.3,89.32,50Zm11.77-3.08-1.4.36-.77.2c-2.19.57-2.68.21-3.62-1.91a42.4,42.4,0,0,0-3.91-7.76c-7.3-10-17-14.34-29.28-12.41-1.47.24-2.36-.22-2.72-1.71-.22-1-.47-1.9-.68-2.86-.38-1.72.17-2.58,1.86-2.89,10.6-2,20.2.43,28.84,6.86,6.46,4.81,10.69,11.29,13.32,19C103.34,45.49,102.85,46.39,101.09,46.89Z" />
			</svg>
  		`
	case "ethermine_staking_logo_svg":
		svg = `
		<svg width="30" height="30" viewBox="0 0 30 30" transform="scale(1.8)" fill="none" xmlns="http://www.w3.org/2000/svg">
			<path d="M15 21.6626L6.6687 16.7563L15 28.4813L23.3312 16.7563L15 21.6626Z" style="fill:var(--body-color-inverted)"/>
			<path d="M15.025 10.2312L18.8937 7.93744L15.025 1.51245L11.1062 8.01869L15.025 10.2312Z" style="fill:var(--body-color-inverted)"/>
			<path d="M14.9999 20.2251L23.3187 15.3376L19.4312 8.88135L14.9999 11.5063L10.5562 8.9001L6.68115 15.3313L14.9999 20.2251Z" style="fill:var(--body-color-inverted)"/>
		</svg>
		`
	case "ethermine_stake_logo_svg":
		svg = `
			<svg viewBox="7.98 -0.02 24.78 40.07">
				<path data-name="Rechteck 13" transform="translate(0 .03)" style="fill:none" d="M0 0h40v40H0z"></path>
				<g data-name="Ebene 2">
				<path data-name="Pfad 25" d="M165.132 200.79v-10.124l12.375-7.286z" transform="translate(-144.758 -160.76)" style="fill:var(--body-color)"></path>
				<path data-name="Linie 4" transform="translate(8 22.62)" style="fill:var(--body-color)" d="m0 0 12.374 17.41"></path>
				<path data-name="Pfad 26" d="M107.764 200.79v-10.124L95.39 183.38" transform="translate(-87.39 -160.76)" style="fill:var(--body-color)"></path>
				<path data-name="Pfad 27" d="M138.035 68.872v-13l-5.845 9.707z" transform="translate(-117.662 -55.87)" style="fill:var(--font-color)"></path>
				<path data-name="Pfad 28" d="m95.52 127.209 12.351 7.263v-12.945l-6.6-3.867z" transform="translate(-87.497 -106.699)" style="fill:var(--font-color)"></path>
				<path data-name="Pfad 29" d="m170.913 65.455-5.773-9.585v13z" transform="translate(-144.766 -55.87)" style="fill:var(--body-color)"></path>
				<path data-name="Pfad 30" d="m171.721 117.48-6.581 3.9v12.945l12.351-7.263z" transform="translate(-144.766 -106.55)" style="fill:var(--body-color)"></path>
				<path data-name="Pfad 31" d="m95.39 145.142 12.351-5.682.021 2.775-.021 10.17z" transform="translate(-87.39 -124.631)" style="fill:var(--body-color)"></path>
				</g>
			</svg>
		`
	case "webhook_logo_svg":
		svg = `
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" style="margin-bottom: 3px; fill: var(--font-color);" xmlns="http://www.w3.org/2000/svg">
				<!-- credit https://www.iconfinder.com/carbon-design -->
				<defs>
				<style>
					.webhook-cls-1 {
					fill: none;
					}
				</style>
				</defs>
				<path d="M24,26a3,3,0,1,0-2.8164-4H13v1a5,5,0,1,1-5-5V16a7,7,0,1,0,6.9287,8h6.2549A2.9914,2.9914,0,0,0,24,26Z" />
				<path d="M24,16a7.024,7.024,0,0,0-2.57.4873l-3.1656-5.5395a3.0469,3.0469,0,1,0-1.7326.9985l4.1189,7.2085.8686-.4976a5.0006,5.0006,0,1,1-1.851,6.8418L17.937,26.501A7.0005,7.0005,0,1,0,24,16Z" />
				<path d="M8.532,20.0537a3.03,3.03,0,1,0,1.7326.9985C11.74,18.47,13.86,14.7607,13.89,14.708l.4976-.8682-.8677-.497a5,5,0,1,1,6.812-1.8438l1.7315,1.002a7.0008,7.0008,0,1,0-10.3462,2.0356c-.457.7427-1.1021,1.8716-2.0737,3.5728Z" />
				<rect class="webhook-cls-1" data-name="&lt;Transparent Rectangle&gt;" height="32" id="_Transparent_Rectangle_" width="32" />
			</svg>
		`
	case "eversteel_logo_svg":
		svg = `
			<svg width="30" height="26" viewBox="0 0 30 26" fill="none" xmlns="http://www.w3.org/2000/svg">
				<g clip-path="url(#clip0_513_4696)">
				<path d="M15.5002 12.0469C16.2002 13.27 16.8728 14.4468 17.6145 15.7451C20.7184 10.4403 23.7514 5.25692 26.8277 0H2.15936C1.45682 1.20493 0.766792 2.38838 0 3.70405H20.3747C18.7109 6.55194 17.1164 9.28081 15.5002 12.0469ZM17.597 17.9244C15.9324 15.0361 14.3521 12.2956 12.7643 9.54031H8.50564C11.5411 14.8146 14.5407 20.0261 17.6003 25.3424C21.7813 18.2153 25.8864 11.218 30 4.20487C29.2924 2.97845 28.6082 1.79335 27.8698 0.514039C24.4256 6.35195 21.0456 12.0808 17.597 17.9244ZM0.0575719 4.77429C4.2111 11.8684 8.30788 18.8658 12.3788 25.821H16.67C13.2766 19.9888 9.92574 14.2303 6.51898 8.3767H16.2603C16.962 7.18995 17.6504 6.02634 18.3905 4.77429H0.0575719Z" style="fill:var(--body-color-inverted)"/>
				</g>
				<defs>
				<clipPath id="clip0_513_4696">
				<rect width="30" height="26" fill="white"/>
				</clipPath>
				</defs>
			</svg>
		`
	}

	return template.HTML(string(svg))
}
