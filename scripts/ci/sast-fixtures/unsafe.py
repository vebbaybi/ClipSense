# These are scanner fixtures, never imported or executed.
def unsafe(value):
    # ruleid: python-dynamic-execution
    return eval(value)
