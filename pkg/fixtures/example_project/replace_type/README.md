## Fix replace-type for different packages from the same source

[Issue 710](https://github.com/vektra/mockery/pull/710)

This package is used to test the case where multiple types come from the same package (`replace_type/rti/internal`),
but results in types in different packages (`replace_type/rt1` and `replace_type/rt2`).

Tests `TestReplaceTypePackageMultiplePrologue` and `TestReplaceTypePackageMultiple` use it to check if this outputs
the correct import and type names.