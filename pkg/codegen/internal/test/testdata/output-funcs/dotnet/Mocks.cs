// Copyright 2016-2021, Pulumi Corporation

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Linq;
using System.Threading.Tasks;

using Pulumi.Testing;

namespace Pulumi.MadeupPackage.Codegentest
{
    public class Mocks : IMocks
    {
        public Task<object> CallAsync(MockCallArgs args)
        {

            if (args.Token == "madeup-package:codegentest:funcWithAllOptionalInputs"
                || args.Token == "madeup-package:codegentest:funcWithDefaultValue")
            {
                var dictBuilder = ImmutableDictionary.CreateBuilder<string,Object>();
                var a = args.Args.GetValueOrDefault("a", "null");
                var b = args.Args.GetValueOrDefault("b", "null");
                dictBuilder.Add("r", $"a={a} b={b}");
                var result = dictBuilder.ToImmutableDictionary();
                return Task.FromResult((Object)result);
            }

            if (args.Token == "madeup-package:codegentest:funcWithDictParam")
            {
                string aString = "null";
                if (args.Args.ContainsKey("a"))
                {
                    var a = (ImmutableDictionary<string,object>)args.Args["a"];
                    aString = string.Join(", ", a.Keys.OrderBy(k => k).Select(k => $"{k}: {a[k]}"));
                }
                var b = args.Args.GetValueOrDefault("b", "null");
                var dictBuilder = ImmutableDictionary.CreateBuilder<string,Object>();
                dictBuilder.Add("r", $"a=[{aString}] b={b}");
                var result = dictBuilder.ToImmutableDictionary();
                return Task.FromResult((Object)result);
            }

            if (args.Token == "madeup-package:codegentest:funcWithListParam")
            {
                string aString = "null";
                if (args.Args.ContainsKey("a"))
                {
                    var a = (ImmutableArray<object>)args.Args["a"];
                    aString = string.Join(", ", a.OrderBy(k => k).Select(e => $"{e}"));
                }
                var b = args.Args.GetValueOrDefault("b", "null");
                var dictBuilder = ImmutableDictionary.CreateBuilder<string,Object>();
                dictBuilder.Add("r", $"a=[{aString}] b={b}");
                var result = dictBuilder.ToImmutableDictionary();
                return Task.FromResult((Object)result);
            }

            throw new Exception($"CallAsync not implemented for {args.Token}..");
        }

        public Task<(string? id, object state)> NewResourceAsync(MockResourceArgs args)
        {
            throw new Exception("NewResourceAsync not implemented..");
        }
    }
}