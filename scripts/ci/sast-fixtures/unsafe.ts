// Scanner fixture, never bundled or executed.
function unsafe(value: string) {
  // ruleid: javascript-dynamic-execution
  return eval(value);
}
