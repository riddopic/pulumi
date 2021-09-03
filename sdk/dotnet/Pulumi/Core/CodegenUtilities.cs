// Copyright 2016-2021, Pulumi Corporation

using System.Collections.Generic;
using System.Diagnostics.CodeAnalysis;

namespace Pulumi.Utilities
{
    /// <summary>
    /// Supports automatically generated Pulumi code, such as
    /// `pulumi-azure-native` provider.
    /// </summary>
    public static class CodegenUtilities
    {
        public static Input<Dictionary<string,T>> ToDict<T>(this InputMap<T> inputMap)
            => inputMap.Apply(v => new Dictionary<string,T>(v));

        public static Input<List<T>> ToList<T>(this InputList<T> inputList)
            => inputList.Apply(v => new List<T>(v));

        public class Boxed
        {
            [AllowNull]
            public object Value { get; }

            private Boxed([AllowNull] object o)
            {
                Value = o;
            }

            public static Boxed Create([AllowNull] object o)
            {
                return new Boxed(o);
            }

            public void Set(object target, string propertyName)
            {
                var v = this.Value;
                if (v != null)
                {
                    var p = target.GetType().GetProperty(propertyName);
                    if (p != null)
                    {
                        p.SetValue(target, v);
                    }
                }
            }
        }

        public static Output<Boxed> Box<T>([AllowNull] this Input<T> input)
        {
            if (input == null)
            {
                return Output.Create(Boxed.Create(null));
            }
            else
            {
                return input.Apply(v => Boxed.Create(v));
            }
        }
    }
}
