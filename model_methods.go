package ffmpegsplit

// NumChapters returns the number of chapters found in the input file.
func (imeta InputFileMetadata) NumChapters() int {
	return len(imeta.FFProbeOutput.Chapters)
}

// AddFilter appends appends a filter to the list of filters in the OutFileOpts struct
func (opts *OutFileOpts) AddFilter(flt ChapterFilter) {
	opts.Filters = append(opts.Filters, flt)
}

// IsFiltered invokes all configured filters for the given Chapter. If any of the
// filters return true, the function returns true. In other words, IsFiltered()
// returns false iff all the filters return false for the chapter.
func (opts OutFileOpts) IsFiltered(ch Chapter) bool {
	for i := range opts.Filters {
		if opts.Filters[i].Filter(ch) {
			return true
		}
	}
	return false
}

// DefaultOutFileOpts returns some sensible set of default values for OutFileOpts.
func DefaultOutFileOpts() OutFileOpts {
	var opts OutFileOpts
	opts.UseTitleInName = true
	opts.UseTitleInMeta = true
	opts.UseChapterNumberInMeta = true
	opts.EnumOffset = -1
	opts.EnumPaddedWidth = -1
	return opts

}
