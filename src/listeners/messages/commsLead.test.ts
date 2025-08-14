import { commsLeadRegex } from "./commsLead";

describe("commsLeadRegex", () => {
  it("should match when tagging flarebot with just 'comms lead'", () => {
    const text = "@somebot comms lead";
    expect(text.match(commsLeadRegex)).toBeTruthy();
  });

  it("should match when just saying 'comms lead'", () => {
    const text = "comms lead";
    expect(text.match(commsLeadRegex)).toBeTruthy();
  });

  it("should match 'i am comms lead'", () => {
    const text = "i am comms lead";
    expect(text.match(commsLeadRegex)).toBeTruthy();
  });

  it("should match 'I am comms lead' (case insensitive)", () => {
    const text = "I am comms lead";
    expect(text.match(commsLeadRegex)).toBeTruthy();
  });

  it("should match 'i'm comms lead'", () => {
    const text = "i'm comms lead";
    expect(text.match(commsLeadRegex)).toBeTruthy();
  });

  it("should match 'I'm comms lead'", () => {
    const text = "I'm comms lead";
    expect(text.match(commsLeadRegex)).toBeTruthy();
  });

  it("should match 'i am the comms lead'", () => {
    const text = "i am the comms lead";
    expect(text.match(commsLeadRegex)).toBeTruthy();
  });

  it("should match 'I am the comms lead'", () => {
    const text = "I am the comms lead";
    expect(text.match(commsLeadRegex)).toBeTruthy();
  });

  it("should match 'i'm the comms lead'", () => {
    const text = "i'm the comms lead";
    expect(text.match(commsLeadRegex)).toBeTruthy();
  });

  it("should not match 'you are comms lead'", () => {
    const text = "you are comms lead";
    expect(text.match(commsLeadRegex)).toBeFalsy();
  });

  it("should not match 'who is comms lead'", () => {
    const text = "who is comms lead";
    expect(text.match(commsLeadRegex)).toBeFalsy();
  });

  it("should match with surrounding text", () => {
    const text = "Hey everyone, i am comms lead for this issue";
    expect(text.match(commsLeadRegex)).toBeTruthy();
  });
});
