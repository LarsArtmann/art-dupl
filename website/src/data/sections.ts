import type { StepCard, ComparisonItem, ComparisonMatrix, UseCase, OutputFormat } from "./types";

export const steps: StepCard[] = [
  {
    step: "1",
    title: "Parse",
    desc: "Scan .go and .templ files, parse into ASTs with concurrent workers.",
    code: "go/parser + templ parser",
  },
  {
    step: "2",
    title: "Serialize",
    desc: "Transform AST into unified token sequences with semantic encoding.",
    code: "syntax.Node[]",
  },
  {
    step: "3",
    title: "Detect",
    desc: "Build compressed suffix tree, find repeating patterns above threshold.",
    code: "Ukkonen O(n)",
  },
  {
    step: "4",
    title: "Classify & Output",
    desc: "Label clone types, score extractability, filter boilerplate, format output.",
    code: "7 output formats",
  },
];

export const comparisons: ComparisonItem[] = [
  {
    variant: "dupl (original)",
    accent: false,
    pros: ["Simple suffix tree", "Text + HTML output"],
    cons: [
      "No semantic matching",
      "No generated-code filtering",
      "No CI gating",
      "No templ support",
      "No stats",
    ],
  },
  {
    variant: "jscpd",
    accent: false,
    pros: ["Multi-language", "Token-based"],
    cons: [
      "Not AST-aware",
      "No Go semantic understanding",
      "No templ support",
      "No CI baseline gating",
    ],
  },
  {
    variant: "art-dupl",
    accent: true,
    pros: [
      "AST-level suffix tree + hash detection",
      "3 matching modes (semantic, exact, structural)",
      "7 output formats including SARIF",
      "Generated-code auto-filtering",
      "CI baseline gating + pre-commit hook",
      "Full templ support",
      "Programmatic SDK",
    ],
    cons: [],
  },
];

export const comparisonMatrix: ComparisonMatrix = {
  columns: ["dupl (original)", "jscpd", "art-dupl"],
  rows: [
    { feature: "AST-level analysis", values: ["yes", "no", "yes"] },
    { feature: "Semantic matching (Type 2)", values: ["no", "no", "yes"] },
    { feature: "Generated-code filtering", values: ["no", "partial", "yes"] },
    { feature: "CI baseline gating", values: ["no", "no", "yes"] },
    { feature: "Templ support", values: ["no", "no", "yes"] },
    { feature: "SARIF output", values: ["no", "no", "yes"] },
    { feature: "Stats + health grades", values: ["no", "partial", "yes"] },
    { feature: "Programmatic SDK", values: ["no", "no", "yes"] },
  ],
};

export const useCases: UseCase[] = [
  {
    title: "Terminal Workflow",
    desc: "Quick scan during development to catch duplication before it spreads",
    icon: "terminal",
  },
  {
    title: "CI/CD Pipeline",
    desc: "Gate merges on new clones with baseline + check workflow",
    icon: "pipeline",
  },
  {
    title: "Code Review",
    desc: "HTML reports with side-by-side diffs for thorough review sessions",
    icon: "review",
  },
  {
    title: "Refactoring",
    desc: "Extractability scores and clone classification prioritize what to fix",
    icon: "refactor",
  },
  {
    title: "Health Monitoring",
    desc: "Stats subcommand tracks duplication trends with A-F health grades",
    icon: "monitor",
  },
];

export const outputFormats: OutputFormat[] = [
  { flag: "(default)", name: "Text", desc: "Quick terminal review" },
  { flag: "--rich-text", name: "Rich Text", desc: "Priority badges and suggestions" },
  { flag: "--html", name: "HTML", desc: "Syntax highlighting and diffs" },
  { flag: "--json", name: "JSON", desc: "CI/CD with metadata and summary" },
  { flag: "--simple-json", name: "Simple JSON", desc: "Lightweight with impact scores" },
  { flag: "--sarif", name: "SARIF", desc: "GitHub Advanced Security" },
  { flag: "--plumbing", name: "Plumbing", desc: "Machine-readable file:line format" },
];
