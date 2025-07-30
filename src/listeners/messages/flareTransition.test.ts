/* eslint-env jest */
import { extractFlareTransition } from "./flareTransition";

describe("extractFlareTransition", () => {
  const testCases = [
    // Positive cases
    { text: "mitigated", expected: "mitigated" },
    { text: "mitigate", expected: "mitigate" },
    { text: "flare mitigated", expected: "mitigated" },
    { text: "flare is mitigated", expected: "mitigated" },
    { text: "unmitigated", expected: "unmitigated" },
    { text: "unmitigate", expected: "unmitigate" },
    { text: "flare unmitigated", expected: "unmitigated" },
    { text: "not a flare", expected: "not a flare" },
    { text: "not flare", expected: "not flare" },
    { text: "flare not a flare", expected: "not a flare" },
    { text: "flare not flare", expected: "not flare" },
    { text: "flare is not a flare", expected: "not a flare" },
    { text: "flare is not flare", expected: "not flare" },
    { text: "MITIGATED", expected: "mitigated" },
    { text: "UNMITIGATED", expected: "unmitigated" },
    { text: "NOT A FLARE", expected: "not a flare" },
    { text: "NOT FLARE", expected: "not flare" },
    { text: "The flare is mitigated now", expected: "mitigated" },
    { text: "Please mitigate the flare", expected: "mitigate" },
    { text: "flare was mitigated", expected: "mitigated" },
    { text: "mitigated flare", expected: "mitigated" },

    // Negative cases
    { text: "flare", expected: null },
    { text: "mitigation", expected: null },
    { text: "unmitigation", expected: null },
    { text: "flare is", expected: null },
    { text: "flare mitigation", expected: null },
  ];

  testCases.forEach(({ text, expected }) => {
    test(`extracts "${expected}" from "${text}"`, () => {
      const result = extractFlareTransition(text);
      expect(result).toBe(expected);
    });
  });
});
