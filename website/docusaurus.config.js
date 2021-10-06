const lightCodeTheme = require("prism-react-renderer/themes/github");
const darkCodeTheme = require("prism-react-renderer/themes/dracula");

// With JSDoc @type annotations, IDEs can provide config autocompletion
/** @type {import('@docusaurus/types').DocusaurusConfig} */
(
  module.exports = {
    title: "Groot",
    tagline: "GraphQL in Go",
    url: "https://shreyas44.github.io",
    baseUrl: "/groot/",
    onBrokenLinks: "throw",
    onBrokenMarkdownLinks: "warn",
    // favicon: 'img/favicon.ico',
    organizationName: "shreyas44",
    projectName: "groot",

    presets: [
      [
        "@docusaurus/preset-classic",
        /** @type {import('@docusaurus/preset-classic').Options} */
        ({
          docs: {
            sidebarPath: require.resolve("./sidebars.js"),
            editUrl: "https://github.com/shreyas44/groot/edit/main/docs/",
          },
        }),
      ],
    ],

    themeConfig:
      /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
      ({
        navbar: {
          title: "Groot",
          items: [
            {
              href: "https://github.com/shreyas44/groot",
              label: "GitHub",
              position: "right",
            },
          ],
        },
        // footer: {
        //   style: "dark",
        //   copyright: `Copyright © ${new Date().getFullYear()} Shreyas Built with Docusaurus.`,
        // },
        prism: {
          theme: lightCodeTheme,
          darkTheme: darkCodeTheme,
        },
      }),
  }
);
