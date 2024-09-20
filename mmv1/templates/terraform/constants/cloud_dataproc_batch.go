/*
 * Dataproc Batch api apends subminor version to the provided
 * version. We are suppressing this server generated subminor.
 */
func CloudDataprocBatchRuntimeConfigVersionDiffSuppressFunc(old, new string) bool {
	if old != "" && strings.HasPrefix(new, old) || (new != "" && strings.HasPrefix(old, new)) {
		return true
	}

	return old == new
}

func CloudDataprocBatchRuntimeConfigVersionDiffSuppress(_, old, new string, d *schema.ResourceData) bool {
	return CloudDataprocBatchRuntimeConfigVersionDiffSuppressFunc(old, new)
}
