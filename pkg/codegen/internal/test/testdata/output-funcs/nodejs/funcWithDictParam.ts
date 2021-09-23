// *** WARNING: this file was generated by test. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import { input as inputs, output as outputs } from "./types";
import * as utilities from "./utilities";

/**
 * Check codegen of functions with a Dict<str,str> parameter.
 */
export function funcWithDictParam(args?: FuncWithDictParamArgs, opts?: pulumi.InvokeOptions): Promise<FuncWithDictParamResult> {
    args = args || {};
    if (!opts) {
        opts = {}
    }

    if (!opts.version) {
        opts.version = utilities.getVersion();
    }
    return pulumi.runtime.invoke("mypkg::funcWithDictParam", {
        "a": args.a,
        "b": args.b,
    }, opts);
}

export interface FuncWithDictParamArgs {
    a?: {[key: string]: string};
    b?: string;
}

export interface FuncWithDictParamResult {
    readonly r: string;
}
