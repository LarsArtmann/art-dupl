import type { Feature } from "./types";

export const features: Feature[] = [
  {
    icon: "tree",
    title: "Suffix Tree Detection",
    desc: "Ukkonen's algorithm on serialized ASTs finds exact and near-miss clones that regex tools miss. O(n) with O(1) map-based transitions.",
  },
  {
    icon: "semantic",
    title: "Semantic Awareness",
    desc: "Alpha-normalization detects renamed-variable clones (Type 2). Three modes: semantic (default), exact, and structural.",
  },
  {
    icon: "output",
    title: "7 Output Formats",
    desc: "Text, rich-text, HTML with diffs, JSON, Simple-JSON, SARIF for GitHub Security, and machine-readable plumbing.",
  },
  {
    icon: "filter",
    title: "Smart Generated-Code Filtering",
    desc: "Auto-detects and filters sqlc, protobuf, mockgen, stringer, templ, and generic generated code. Override per-category.",
  },
  {
    icon: "ci",
    title: "CI/CD Baseline Gating",
    desc: "Record accepted clones, then gate on new ones. Pre-commit hook, GitHub Actions template, and SARIF upload included.",
  },
  {
    icon: "sdk",
    title: "Programmatic SDK",
    desc: "Detector interface with streaming, progress callbacks, and 20+ sentinel errors. Zero config imports — fully independent types.",
  },
];
