import { defineConfig, fontProviders } from "astro/config";
import starlight from "@astrojs/starlight";
import sitemap from "@astrojs/sitemap";

import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
	site: "https://art-dupl.lars.software",

	compressHTML: true,

	prefetch: {
		prefetchAll: false,
		defaultStrategy: "hover",
	},

	fonts: [
		{
			provider: fontProviders.google(),
			name: "Syne",
			cssVariable: "--font-syne",
			weights: [400, 500, 600, 700, 800],
			styles: ["normal"],
			subsets: ["latin"],
			fallbacks: ["sans-serif"],
		},
		{
			provider: fontProviders.fontsource(),
			name: "JetBrains Mono",
			cssVariable: "--font-jetbrains-mono",
			weights: [400, 500, 600, 700],
			styles: ["normal"],
			subsets: ["latin"],
			fallbacks: ["monospace"],
		},
	],

	integrations: [
		sitemap(),
		starlight({
			title: "art-dupl",
			favicon: "/favicon.svg",
			customCss: ["./src/styles/starlight.css"],
			expressiveCode: {
				themes: ["github-light", "github-dark"],
				frames: {
					showCopyToClipboardButton: true,
				},
			},
			sidebar: [
				{
					label: "Getting Started",
					items: [
						{ label: "Installation", slug: "getting-started/installation" },
						{ label: "Quick Start", slug: "getting-started/quick-start" },
					],
				},
				{
					label: "Guides",
					items: [
						{ label: "Detection Methods", slug: "guides/detection-methods" },
						{ label: "Output Formats", slug: "guides/output-formats" },
						{ label: "CI/CD Integration", slug: "guides/ci-cd" },
						{ label: "Smart Filtering", slug: "guides/filtering" },
						{ label: "Performance", slug: "guides/performance" },
						{ label: "SDK", slug: "guides/sdk" },
					],
				},
				{
					label: "Reference",
					items: [
						{ label: "Configuration", slug: "configuration" },
						{ label: "CLI Flags", slug: "cli-flags" },
					],
				},
				{
					label: "Community",
					items: [
						{ label: "Changelog", slug: "changelog" },
						{ label: "Contributing", slug: "contributing" },
						{ label: "Related Tools", slug: "related-tools" },
					],
				},
			],
			social: [
				{
					icon: "github",
					label: "GitHub",
					href: "https://github.com/LarsArtmann/art-dupl",
				},
			],
			head: [
				{
					tag: "meta",
					attrs: {
						name: "description",
						content:
							"Professional code clone detection for Go. AST-based structural detection with suffix tree algorithms, semantic awareness, and 7 output formats.",
					},
				},
			],
		}),
	],

	vite: {
		plugins: [tailwindcss()],
	},
});
