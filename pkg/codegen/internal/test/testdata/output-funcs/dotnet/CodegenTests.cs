// Copyright 2016-2021, Pulumi Corporation

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

using FluentAssertions;
using NUnit.Framework;
using Moq;

using Pulumi;
using Pulumi.Testing;
using static Pulumi.MadeupPackage.Codegentest.TestHelpers;

using Pulumi.MadeupPackage.Codegentest.Outputs;

namespace Pulumi.MadeupPackage.Codegentest
{
    [TestFixture]
    public class UnitTests
    {
        [Test]
        public async Task FuncWithAllOptionalInputsOutputWorks()
        {
            Func<string,Func<FuncWithAllOptionalInputsInputArgs?>,Task> check = (
                (expected, args) => Assert
                .Output(() => FuncWithAllOptionalInputs.Invoke(args()).Apply(x => x.R))
                .ResolvesTo(expected)
            );

            await check("a=null b=null", () => null);

            await check("a=null b=null", () => new FuncWithAllOptionalInputsInputArgs());

            await check("a=my-a b=null", () => new FuncWithAllOptionalInputsInputArgs
            {
                A = Out("my-a"),
            });

            await check("a=null b=my-b", () => new FuncWithAllOptionalInputsInputArgs
            {
                B = Out("my-b"),
            });

            await check("a=my-a b=my-b", () => new FuncWithAllOptionalInputsInputArgs
            {
                A = Out("my-a"),
                B = Out("my-b"),
            });
        }

        [Test]
        public async Task FuncWithDefaultValueOutputWorks()
        {
            Func<string,Func<FuncWithDefaultValueInputArgs>,Task> check = (
                (expected, args) => Assert
                .Output(() => FuncWithDefaultValue.Invoke(args()).Apply(x => x.R))
                .ResolvesTo(expected)
            );

            // Since A is required, not passing it is an exception.
            Func<Task> act = () => check("", () => new FuncWithDefaultValueInputArgs());
            await act.Should().ThrowAsync<Exception>();

            // Check that default values from the schema work.
            await check("a=my-a b=b-default", () => new FuncWithDefaultValueInputArgs()
            {
                A = Out("my-a")
            });

            await check("a=my-a b=my-b", () => new FuncWithDefaultValueInputArgs()
            {
                A = Out("my-a"),
                B = Out("my-b")
            });
        }

        [Test]
        public async Task FuncWithDictParamOutputWorks()
        {
            Func<string,Func<FuncWithDictParamInputArgs>,Task> check = (
                (expected, args) => Assert
                .Output(() => FuncWithDictParam.Invoke(args()).Apply(x => x.R))
                .ResolvesTo(expected)
            );

            var map = new InputMap<string>();
            map.Add("K1", Out("my-k1"));
            map.Add("K2", Out("my-k2"));

            // Omitted value defaults to empty dict and not null.
            await check("a=[] b=null", () => new FuncWithDictParamInputArgs());

            await check("a=[K1: my-k1, K2: my-k2] b=null", () => new FuncWithDictParamInputArgs()
            {
                A = map,
            });

            await check("a=[K1: my-k1, K2: my-k2] b=my-b", () => new FuncWithDictParamInputArgs()
            {
                A = map,
                B = Out("my-b"),
            });
        }

        [Test]
        public async Task FuncWithListParamOutputWorks()
        {
            Func<string,Func<FuncWithListParamInputArgs>,Task> check = (
                (expected, args) => Assert
                .Output(() => FuncWithListParam.Invoke(args()).Apply(x => x.R))
                .ResolvesTo(expected)
            );

            var lst = new InputList<string>();
            lst.Add("e1");
            lst.Add("e2");
            lst.Add("e3");

            // Similarly to dicts, omitted value defaults to empty list and not null.
            await check("a=[] b=null", () => new FuncWithListParamInputArgs());

            await check("a=[e1, e2, e3] b=null", () => new FuncWithListParamInputArgs()
            {
                A = lst,
            });

            await check("a=[e1, e2, e3] b=my-b", () => new FuncWithListParamInputArgs()
            {
                A = lst,
                B = Out("my-b"),
            });
        }

        [Test]
        public async Task GetIntegrationRuntimeObjectMetadatumOuputWorks()
        {
            Func<string,Func<GetIntegrationRuntimeObjectMetadatumInputArgs>,Task> check = (
                (expected, args) => Assert
                .Output(() => GetIntegrationRuntimeObjectMetadatum.Invoke(args()).Apply(x => {
                    var nextLink = x.NextLink ?? "null";
                    var valueRepr = "null";
                    if (x.Value != null)
                    {
                        valueRepr = string.Join(", ", x.Value.Select(e => $"{e}"));
                    }
                    return $"nextLink={nextLink} value=[{valueRepr}]";
                 }))
                .ResolvesTo(expected)
            );

            await check("nextLink=my-next-link value=[factoryName: my-fn, integrationRuntimeName: my-irn, " +
                        "metadataPath: my-mp, resourceGroupName: my-rgn]",
                        () => new GetIntegrationRuntimeObjectMetadatumInputArgs()
                        {
                            FactoryName = Out("my-fn"),
                            IntegrationRuntimeName = Out("my-irn"),
                            MetadataPath = Out("my-mp"),
                            ResourceGroupName = Out("my-rgn")
                        });
        }

        [Test]
        public async Task TestListStorageAccountsOutputWorks()
        {
            Func<StorageAccountKeyResponse, string> showSAKR = (r) =>
                $"CreationTime={r.CreationTime} KeyName={r.KeyName} Permissions={r.Permissions} Value={r.Value}";

            Func<string,Func<ListStorageAccountKeysInputArgs>,Task> check = (
                (expected, args) => Assert
                .Output(() => ListStorageAccountKeys.Invoke(args()).Apply(x => {
                    return "[" + string.Join(", ", x.Keys.Select(k => showSAKR(k))) + "]";
                })).ResolvesTo(expected)
            );

            await check("[CreationTime=my-creation-time KeyName=my-key-name Permissions=my-permissions" +
                        " Value=[accountName: my-an, expand: my-expand, resourceGroupName: my-rgn]]",
                        () => new ListStorageAccountKeysInputArgs()
                        {
                            AccountName = Out("my-an"),
                            ResourceGroupName = Out("my-rgn"),
                            Expand = Out("my-expand")
                        });
        }
    }
}
