// *** WARNING: this file was generated by test. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.Example
{
    [ExampleResourceType("example::ToyStore")]
    public partial class ToyStore : Pulumi.CustomResource
    {
        [Output("chew")]
        public Output<Pulumi.Example.Chew?> Chew { get; private set; } = null!;

        [Output("laser")]
        public Output<Pulumi.Example.Laser?> Laser { get; private set; } = null!;

        [Output("stuff")]
        public Output<ImmutableArray<Outputs.Toy>> Stuff { get; private set; } = null!;

        [Output("wanted")]
        public Output<ImmutableArray<Outputs.Toy>> Wanted { get; private set; } = null!;


        /// <summary>
        /// Create a ToyStore resource with the given unique name, arguments, and options.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resource</param>
        /// <param name="args">The arguments used to populate this resource's properties</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public ToyStore(string name, ToyStoreArgs? args = null, CustomResourceOptions? options = null)
            : base("example::ToyStore", name, args ?? new ToyStoreArgs(), MakeResourceOptions(options, ""))
        {
        }

        private ToyStore(string name, Input<string> id, CustomResourceOptions? options = null)
            : base("example::ToyStore", name, null, MakeResourceOptions(options, id))
        {
        }

        private static CustomResourceOptions MakeResourceOptions(CustomResourceOptions? options, Input<string>? id)
        {
            var defaultOptions = new CustomResourceOptions
            {
                Version = Utilities.Version,
                ReplaceOnChanges =
                {
                    "stuff[*].associated.color",
                    "stuff[*].color",
                    "wanted[*].associated.color",
                    "wanted[*].color",
                },
            };
            var merged = CustomResourceOptions.Merge(defaultOptions, options);
            // Override the ID if one was specified for consistency with other language SDKs.
            merged.Id = id ?? merged.Id;
            return merged;
        }
        /// <summary>
        /// Get an existing ToyStore resource's state with the given name, ID, and optional extra
        /// properties used to qualify the lookup.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resulting resource.</param>
        /// <param name="id">The unique provider ID of the resource to lookup.</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public static ToyStore Get(string name, Input<string> id, CustomResourceOptions? options = null)
        {
            return new ToyStore(name, id, options);
        }
    }

    public sealed class ToyStoreArgs : Pulumi.ResourceArgs
    {
        public ToyStoreArgs()
        {
        }
    }
}
