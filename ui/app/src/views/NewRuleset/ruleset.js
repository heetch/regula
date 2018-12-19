const defaultRule = '(#true)';

// Describes rule S-Expression and return value.
class Rule {
  constructor(sExpr = defaultRule, returnValue = '') {
    this.sExpr = sExpr;
    this.returnValue = returnValue;
  }
}

// Describes a ruleset parameter.
class Param {
  constructor(name = '', type = '') {
    this.name = name;
    this.type = type;
  }
}

// Describes a ruleset signature.
class Signature {
  constructor(returnType = '', params = []) {
    this.returnType = returnType;
    this.params = params;
  }
}

// Describes the ruleset payload sent to the server when creating a ruleset.
class Ruleset {
  constructor({ path = '', signature = new Signature(), rules = [] }) {
    this.path = path;
    this.signature = signature;
    this.rules = rules;
  }
}

export { Signature, Param, Ruleset, Rule };

