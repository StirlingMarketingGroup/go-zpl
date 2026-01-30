---
name: visual-diff
description: Compare two images using Gemini's visual understanding to identify differences. Use when comparing rendered output against reference images, visual regression testing, or analyzing what's different between two versions of an image. Invoke with /visual-diff <reference-image> <actual-image>
---

# Visual Diff with Gemini

Compare two images using Gemini 3 Pro's visual understanding.

## Invocation

Parse arguments: `/visual-diff <reference-path> <actual-path>`

## Steps

1. Verify both paths exist (use ls or Read tool)
2. Run gemini comparison:

```bash
gemini --yolo --model gemini-3-pro-preview --output-format text \
  "Compare these two images. First is REFERENCE (expected), second is ACTUAL (rendered).

Report:
1. Overall Match: identical/minor/moderate/major difference?
2. Missing Elements: in reference but not actual?
3. Extra Elements: in actual but not reference?
4. Position Differences: same element, different location?
5. Size/Scale Differences: wrong size?
6. Text Differences: missing, wrong, or different text?
7. Specific Coordinates: approximate pixel locations of issues.

Be specific. Focus on what needs fixing.

Reference: <reference-path>
Actual: <actual-path>"
```

3. Summarize findings to user
