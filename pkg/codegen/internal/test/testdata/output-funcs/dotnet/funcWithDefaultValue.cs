// *** WARNING: this file was generated by . ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;
using Pulumi.Utilities;

namespace Pulumi.MadeupPackage.Codegentest
{
    public static class FuncWithDefaultValue
    {
        /// <summary>
        /// Check codegen of functions with default values.
        /// </summary>
        public static Task<FuncWithDefaultValueResult> InvokeAsync(FuncWithDefaultValueArgs args, InvokeOptions? options = null)
            => Pulumi.Deployment.Instance.InvokeAsync<FuncWithDefaultValueResult>("madeup-package:codegentest:funcWithDefaultValue", args ?? new FuncWithDefaultValueArgs(), options.WithVersion());

        /// <summary>
        /// Check codegen of functions with default values.
        /// </summary>
        public static Output<FuncWithDefaultValueResult> Invoke(FuncWithDefaultValueInvokeArgs args, InvokeOptions? options = null)
        {
            return Pulumi.Output.All(
                args.A.Box(),
                args.B.Box()
            ).Apply(a =>
            {
                var args = new FuncWithDefaultValueArgs();
                a[0].Set(args, nameof(args.A));
                a[1].Set(args, nameof(args.B));
                return InvokeAsync(args, options);
            });
        }
    }


    public sealed class FuncWithDefaultValueArgs : Pulumi.InvokeArgs
    {
        [Input("a", required: true)]
        public string A { get; set; } = null!;

        [Input("b")]
        public string? B { get; set; }

        public FuncWithDefaultValueArgs()
        {
            B = "b-default";
        }
    }

    public sealed class FuncWithDefaultValueInvokeArgs
    {
        public Input<string> A { get; set; } = null!;

        public Input<string>? B { get; set; }

        public FuncWithDefaultValueInvokeArgs()
        {
            B = "b-default";
        }
    }


    [OutputType]
    public sealed class FuncWithDefaultValueResult
    {
        public readonly string R;

        [OutputConstructor]
        private FuncWithDefaultValueResult(string r)
        {
            R = r;
        }
    }
}
