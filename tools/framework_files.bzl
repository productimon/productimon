def _remove_prefix(text, prefix):
    return text[text.startswith(prefix) and len(prefix) + 1:]

# TODO don't copy the headers to reduce size
# and only copy the dylib file in the framework dir not the one in Versions
def _framework_files_impl(ctx):
    all_outputs = []
    for target in ctx.attr.srcs:
        for input in target.files.to_list():
            output_path = _remove_prefix(input.path, target.label.workspace_root)
            output = ctx.actions.declare_file(output_path)

            # print("cp -R", input.path, output.path)
            all_outputs.append(output)
            ctx.actions.run_shell(
                inputs = [input],
                outputs = [output],
                mnemonic = "CopyFrameworkFiles",
                arguments = ["-R", input.path, output.path],
                command = "cp $1 $2 $3",
            )

    return [DefaultInfo(files = depset(all_outputs))]

framework_files = rule(
    implementation = _framework_files_impl,
    attrs = {
        "srcs": attr.label_list(mandatory = True, allow_files = True),
    },
)
