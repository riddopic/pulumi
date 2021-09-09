### Improvements

- [codegen/dotnet] - Add helper function forms `$fn.Invoke` that
  accept `Input`s, return an `Output`, and wrap the underlying
  `$fn.InvokeAsync` call. This change addreses
  [#5758](https://github.com/pulumi/pulumi/issues/) for .NET, making
  it easier to compose functions/datasources with Pulumi resources.
  [#7785](https://github.com/pulumi/pulumi/pull/7855)


### Bug Fixes
