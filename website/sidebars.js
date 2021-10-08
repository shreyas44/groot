/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

module.exports = {
  // By default, Docusaurus generates a sidebar from the docs folder structure
  // mySidebar: [{ type: "autogenerated", dirName: "docs" }],

  // But you can create a sidebar manually
  docs: [
    "introduction",
    "getting-started",
    "context",
    {
      type: "category",
      label: "Type Definitions",
      collapsed: false,
      items: [
        "type-definitions/field-definitions",
        "type-definitions/field-resolvers",
        "type-definitions/object",
        "type-definitions/input",
        "type-definitions/interface",
        "type-definitions/union",
        "type-definitions/enum",
        "type-definitions/scalar",
      ],
    },
    "subscriptions",
    "comparison",
    "composition",
    "relay",
    "internals",
  ],
};
