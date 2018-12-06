package rule

// The rule package defines the mechanism for defining Regula Rules in
// Go code, and for building those Rules via other means (e.g. JSON or
// Regula's symbolic expression language ).
//
// The intent is that number of operations and fundamental types will
// slowly expand, but more importantly that is should be an extensible
// platform.
//
// Currently adding types or operators can only be done by directly
// editing the source of the rules package.
//
// Adding a new operator.
//
// Each new operator must satisfy the Operator interface.  Here is a fully worked example of the "+" operation. For a more complete tour of what is possible I'd suggest reading the definitions of the existing expressions in the expr.go file, once you've read through this example.
//
// In the file expr.go we would define a new struct that implements the Operator interface:
//
//   type exprPlus struct {
//          operator
//   }
//
// We now need a constructor for our new type.  The main role of this
// constructor is to initialise the new structs embedded operator
// struct with a Contract. The Contract allows Regula to introspect
// the Operator, and enforce the type of parameters passed to it.
//
// In this case our operator will have the "OpCode" "plus", it will
// return a NUMBER (a special type that can be either float64 or
// int64).  The arguments we'll pass to "plus" will also be the Type
// NUMBER, and we'll indicate that we'll accept many of them, but that
// we want a minimum of 2 of them.
//
//   func newExprPlus() *exprPlus {
//          return &exprPlus{
//                  operator: operator{
//                          contract: Contract{
//                                  OpCode:     "plus",
//                                  ReturnType: NUMBER,
//                                  Terms:      []Term{{
//                                          Type: NUMBER,
//                                          Cardinality: MANY,
//                                          Min: 2,
//                                  }},
//                          },
//                   },
//           }
//    }
//
// Now we have a public way to construct a "plus" operator, we need to
// define a public function that will insert a "plus" operator into
// the rule's abstract syntax tree.  The function itself should
// reflect the contract we defined.  In this case it's called "Plus",
// and it accepts many Expr arguments, we don't need to worry about
// what Type these expressions are, or exactly how many of them there
// are because we conveyed that information in the contract, and the
// operator.consumeOperands function we use to push arguments onto the
// operation will enforce the contract as we go.  The return type is
// also Expr, because the rule's abstract syntax tree will be a
// homogeneous tree of Exprs.
//
// 🛈 Every one of these public constructors will end up looking the
// same, with only the private constructor they use varying.  This is
// probably a candidate for code generation in the future.
//
//    func Plus(vN ...Expr) Expr {
//            e := newExprPlus()
//            e.consumeOperands(vN...)
//            return e
//    }
//
// To make the operator do anything at all, we have to define its
// evaluator. Our "plus" operator has to add things together!  Our
// Contract says we'll add NUMBER types, which resolve to either
// Float64Value or Int64Value in Regula.  We specifically don't have
// to worry about mixed addition of these types because our Contract
// says we want MANY of type NUMBER, and we can rely on the parser to
// promote INTEGER to FLOAT if we pass a mixed list of NUMBER.  To
// simplify the code for the two concrete types, I've split them out
// of the public Eval method.
//
//    func (n *exprPlus) int64Sum(params) (*Value, error) {
//            var sum int64
//
//            for _, op := range n.operands {
//                    val, err := op.Eval(params)
//                    if err != nil {
//                            return nil, err
//                    }
//                    ival, err != strconv.Atoi(val)
//                    if err != nil {
//                            return nil, err
//                    }
//                    sum += ival
//            }
//            return Int64Value(sum), nil
//    }
//
//    func (n *exprPlus) float64Sum(params) (*Value, error) {
//            var sum float64
//
//            for _, op := range n.operands {
//                    val, err := op.Eval(params)
//                    if err != nil {
//                            return nil, err
//                    }
//                    fval, err != strconv.ParseFloat(val, 64)
//                    if err != nil {
//                            return nil, err
//                    }
//                    sum += fval
//            }
//            return Int64Value(sum), nil
//    }
//
//    func (n *exprPlus) Eval(params Params) (*Value, error) {
//            typ := n.operands[0].Type
//            switch typ {
//            case "int64":
//                    return n.int64Sum(params)
//            case "float64":
//                    return n.float64Sum(params)
//            }
//            return nil, fmt.Errorf("Invalid NUMBER type %q", typ)
//    }
//
//
// Finally, we need a way for people using language other than Go to
// construct our Operator.  The Operator interface defines most of
// what we need, bus we'll need to modify the GetOperator function in
// operator.go to reflect our new operators existence.
//
// For our example, here's GetOperator redefined to include the "plus"
// expression:
//
//   func GetOperator(name string) (Operator, error) {
//        switch name {
//   	  case "eq":
//   	          return newExprEq(), nil
//   	  case "not":
//   	          return newExprNot(), nil
//   	  case "and":
//   		  return newExprAnd(), nil
//   	  case "or":
//   		  return newExprOr(), nil
//   	  case "in":
//   		  return newExprIn(), nil
//   	  case "plus":
//   		  return newExprPlus(), nil
//   	  }
//   	  return nil, fmt.Errorf("no operator Expression called %q exists", name)
//   }
