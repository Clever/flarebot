import { incidentLeadRegex } from "./incidentLead";

describe("incidentLeadRegex", () => {
  it("should match 'i am incident lead'", () => {
    const text = "i am incident lead";
    expect(text.match(incidentLeadRegex)).toBeTruthy();
  });

  it("should match 'incident lead' (case insensitive)", () => {
    const text = "incident lead";
    expect(text.match(incidentLeadRegex)).toBeTruthy();
  });

  it("should match 'I am incident lead' (case insensitive)", () => {
    const text = "I am incident lead";
    expect(text.match(incidentLeadRegex)).toBeTruthy();
  });

  it("should match 'i'm incident lead'", () => {
    const text = "i'm incident lead";
    expect(text.match(incidentLeadRegex)).toBeTruthy();
  });

  it("should match 'I'm incident lead'", () => {
    const text = "I'm incident lead";
    expect(text.match(incidentLeadRegex)).toBeTruthy();
  });

  it("should match 'i am the incident lead'", () => {
    const text = "i am the incident lead";
    expect(text.match(incidentLeadRegex)).toBeTruthy();
  });

  it("should match 'I am the incident lead'", () => {
    const text = "I am the incident lead";
    expect(text.match(incidentLeadRegex)).toBeTruthy();
  });

  it("should match 'i'm the incident lead'", () => {
    const text = "i'm the incident lead";
    expect(text.match(incidentLeadRegex)).toBeTruthy();
  });

  it("should not match 'you are incident lead'", () => {
    const text = "you are incident lead";
    expect(text.match(incidentLeadRegex)).toBeFalsy();
  });

  it("should not match 'who is incident lead'", () => {
    const text = "who is incident lead";
    expect(text.match(incidentLeadRegex)).toBeFalsy();
  });

  it("should match with surrounding text", () => {
    const text = "Hey everyone, i am incident lead for this issue";
    expect(text.match(incidentLeadRegex)).toBeTruthy();
  });
});
