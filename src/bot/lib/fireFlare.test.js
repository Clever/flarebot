/* eslint-env jest */
const { extractPriorityAndTitle } = require("./fireFlare");

describe("extractPriorityAndTitle", () => {
  const cases = [
    {
      text: "fire a flare p0 Something broke",
      expected: {
        priority: "0",
        title: "Something broke",
        specialType: null,
      },
    },
    {
      text: "fire a p1 flare Something else",
      expected: {
        priority: "1",
        title: "Something else",
        specialType: null,
      },
    },
    {
      text: "fire flare p2 Another p2 issue",
      expected: {
        priority: "2",
        title: "Another p2 issue",
        specialType: null,
      },
    },
    {
      text: "fire p0 flare Yet another issue",
      expected: {
        priority: "0",
        title: "Yet another issue",
        specialType: null,
      },
    },
    {
      text: "fire p1 Something urgent",
      expected: {
        priority: "1",
        title: "Something urgent",
        specialType: null,
      },
    },
    {
      text: "fire    a    flare   p2   flare   Lots of spaces",
      expected: {
        priority: "2",
        title: "Lots of spaces",
        specialType: null,
      },
    },
    {
      text: "FIRE A FLARE P0 UPPERCASE",
      expected: {
        priority: "0",
        title: "UPPERCASE",
        specialType: null,
      },
    },
    {
      text: "fire a pre-emptive flare p0 Something broke",
      expected: {
        priority: "0",
        title: "Something broke",
        specialType: "preemptive",
      },
    },
    {
      text: "fire a retroactive p1 flare Something else",
      expected: {
        priority: "1",
        title: "Something else",
        specialType: "retroactive",
      },
    },
    {
      text: "fire a flare preemptive p0 Something broke",
      expected: {
        priority: "0",
        title: "Something broke",
        specialType: "preemptive",
      },
    },
    {
      text: "fire a p2 preemptive flare everything could go down",
      expected: {
        priority: "2",
        title: "everything could go down",
        specialType: "preemptive",
      },
    },
    {
      text: "fire p3 Not a valid priority",
      expected: null,
    },
    {
      text: "fire flare p1",
      expected: null,
    },
    {
      text: "fire a retroactive flare not allowed without priority",
      expected: null,
    },
  ];

  cases.forEach(({ text, expected }) => {
    test(`parses "${text}"`, () => {
      const result = extractPriorityAndTitle(text);
      if (!expected) {
        expect(result).toBeNull();
      } else {
        expect(result).not.toBeNull();
        expect(result.specialType).toBe(expected.specialType);
        expect(result.priority).toBe(expected.priority);
        expect(result.title).toBe(expected.title);
      }
    });
  });
});
