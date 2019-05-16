package etcd

// List returns all rulesets entries or not depending on the query string.
// func TestList(t *testing.T) {
// 	t.Parallel()

// 	s, cleanup := newEtcdRulesetService(t)
// 	defer cleanup()

// 	rsTrue := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}
// 	rsFalse := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(false))}

	// // Root tests the basic behaviour without prefix.
	// t.Run("Root", func(t *testing.T) {
	// 	prefix := "list/root/"

	// 	createBoolRuleset(t, s, prefix+"c", rsTrue...)
	// 	createBoolRuleset(t, s, prefix+"a", rsTrue...)
	// 	createBoolRuleset(t, s, prefix+"a/1", rsTrue...)
	// 	createBoolRuleset(t, s, prefix+"b", rsTrue...)
	// 	createBoolRuleset(t, s, prefix+"a", rsFalse...)

	// 	paths := []string{prefix + "a", prefix + "a/1", prefix + "b", prefix + "c"}

	// 	entries, err := s.List(context.Background(), prefix+"", &api.ListOptions{})
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, len(paths))
	// 	for i, e := range entries.Rulesets {
	// 		require.Equal(t, paths[i], e.Path)
	// 	}
	// 	require.NotEmpty(t, entries.Revision)
	// })

	// // Assert that only latest version for each ruleset is returned by default.
	// t.Run("Last version only", func(t *testing.T) {
	// 	prefix := "list/last/version/"
	// 	rules1 := []*rule.Rule{rule.New(rule.Eq(rule.BoolValue(true), rule.BoolValue(true)), rule.BoolValue(true))}
	// 	rules2 := []*rule.Rule{rule.New(rule.Eq(rule.StringValue("true"), rule.StringValue("true")), rule.BoolValue(true))}

	// 	createBoolRuleset(t, s, prefix+"a", rs...)
	// 	createBoolRuleset(t, s, prefix+"a/1", rs...)
	// 	createBoolRuleset(t, s, prefix+"a", rules1...)
	// 	createBoolRuleset(t, s, prefix+"a", rules2...)

	// 	entries, err := s.List(context.Background(), prefix+"", &api.ListOptions{})
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, 2)
	// 	a := entries.Rulesets[0]
	// 	require.Equal(t, rules2, a.Rules)
	// 	require.NotEmpty(t, entries.Revision)
	// })

	// // Assert that all versions are returned when passing the AllVersions option.
	// t.Run("All versions", func(t *testing.T) {
	// 	prefix := "list/all/version/"
	// 	rules1 := []*rule.Rule{rule.New(rule.Eq(rule.BoolValue(true), rule.BoolValue(true)), rule.BoolValue(true))}
	// 	rules2 := []*rule.Rule{rule.New(rule.Eq(rule.StringValue("true"), rule.StringValue("true")), rule.BoolValue(true))}

	// 	createBoolRuleset(t, s, prefix+"a", rs...)
	// 	time.Sleep(time.Second)
	// 	createBoolRuleset(t, s, prefix+"a", rules1...)
	// 	time.Sleep(time.Second)
	// 	createBoolRuleset(t, s, prefix+"a", rules2...)
	// 	createBoolRuleset(t, s, prefix+"a/1", rs...)

	// 	paths := []string{prefix + "a", prefix + "a", prefix + "a", prefix + "a/1"}

	// 	entries, err := s.List(context.Background(), prefix+"", &api.ListOptions{AllVersions: true})
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, len(paths))
	// 	for i, e := range entries.Rulesets {
	// 		require.Equal(t, paths[i], e.Path)
	// 	}
	// 	require.NotEmpty(t, entries.Revision)

	// 	// Assert that pagination is working well.
	// 	opt := api.ListOptions{
	// 		AllVersions: true,
	// 		Limit:       2,
	// 	}
	// 	entries, err = s.List(context.Background(), prefix+"", &opt)
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, opt.Limit)
	// 	require.Equal(t, prefix+"a", entries.Rulesets[0].Path)
	// 	require.Equal(t, rs, entries.Rulesets[0].Rules)
	// 	require.Equal(t, prefix+"a", entries.Rulesets[1].Path)
	// 	require.Equal(t, rules1, entries.Rulesets[1].Rules)
	// 	require.NotEmpty(t, entries.Revision)

	// 	opt.ContinueToken = entries.Continue
	// 	entries, err = s.List(context.Background(), prefix+"", &opt)
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, opt.Limit)
	// 	require.Equal(t, prefix+"a", entries.Rulesets[0].Path)
	// 	require.Equal(t, rules2, entries.Rulesets[0].Rules)
	// 	require.Equal(t, prefix+"a/1", entries.Rulesets[1].Path)
	// 	require.Equal(t, rs, entries.Rulesets[1].Rules)
	// 	require.NotEmpty(t, entries.Revision)

	// 	t.Run("NotFound", func(t *testing.T) {
	// 		_, err = s.List(context.Background(), prefix+"doesntexist", &api.ListOptions{AllVersions: true})
	// 		require.Equal(t, err, api.ErrRulesetNotFound)
	// 	})

	// })

	// // Prefix tests List with a given prefix.
	// t.Run("Prefix", func(t *testing.T) {
	// 	prefix := "list/prefix/"

	// 	createBoolRuleset(t, s, prefix+"x", rs...)
	// 	createBoolRuleset(t, s, prefix+"xx", rs...)
	// 	createBoolRuleset(t, s, prefix+"x/1", rs...)
	// 	createBoolRuleset(t, s, prefix+"x/2", rs...)

	// 	paths := []string{prefix + "x", prefix + "x/1", prefix + "x/2", prefix + "xx"}

	// 	entries, err := s.List(context.Background(), prefix+"x", &api.ListOptions{})
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, len(paths))
	// 	for i, e := range entries.Rulesets {
	// 		require.Equal(t, paths[i], e.Path)
	// 	}
	// 	require.NotEmpty(t, entries.Revision)
	// })

	// // NotFound tests List with a prefix which doesn't exist.
	// t.Run("NotFound", func(t *testing.T) {
	// 	_, err := s.List(context.Background(), "doesntexist", &api.ListOptions{})
	// 	require.Equal(t, err, api.ErrRulesetNotFound)
	// })

	// // Paging tests List with pagination.
	// t.Run("Paging", func(t *testing.T) {
	// 	prefix := "list/paging/"

	// 	createBoolRuleset(t, s, prefix+"y", rs...)
	// 	createBoolRuleset(t, s, prefix+"yy", rs...)
	// 	createBoolRuleset(t, s, prefix+"y/1", rs...)
	// 	createBoolRuleset(t, s, prefix+"y/2", rs...)
	// 	createBoolRuleset(t, s, prefix+"y/3", rs...)

	// 	opt := api.ListOptions{Limit: 2}
	// 	entries, err := s.List(context.Background(), prefix+"y", &opt)
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, 2)
	// 	require.Equal(t, prefix+"y", entries.Rulesets[0].Path)
	// 	require.Equal(t, prefix+"y/1", entries.Rulesets[1].Path)
	// 	require.NotEmpty(t, entries.Continue)

	// 	opt.ContinueToken = entries.Continue
	// 	token := entries.Continue
	// 	entries, err = s.List(context.Background(), prefix+"y", &opt)
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, 2)
	// 	require.Equal(t, prefix+"y/2", entries.Rulesets[0].Path)
	// 	require.Equal(t, prefix+"y/3", entries.Rulesets[1].Path)
	// 	require.NotEmpty(t, entries.Continue)

	// 	opt.ContinueToken = entries.Continue
	// 	entries, err = s.List(context.Background(), prefix+"y", &opt)
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, 1)
	// 	require.Equal(t, prefix+"yy", entries.Rulesets[0].Path)
	// 	require.Empty(t, entries.Continue)

	// 	opt.Limit = 3
	// 	opt.ContinueToken = token
	// 	entries, err = s.List(context.Background(), prefix+"y", &opt)
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, 3)
	// 	require.Equal(t, prefix+"y/2", entries.Rulesets[0].Path)
	// 	require.Equal(t, prefix+"y/3", entries.Rulesets[1].Path)
	// 	require.Equal(t, prefix+"yy", entries.Rulesets[2].Path)
	// 	require.Empty(t, entries.Continue)

	// 	opt.ContinueToken = "some token"
	// 	entries, err = s.List(context.Background(), prefix+"y", &opt)
	// 	require.Equal(t, api.ErrInvalidContinueToken, err)

	// 	opt.Limit = -10
	// 	opt.ContinueToken = ""
	// 	entries, err = s.List(context.Background(), prefix+"y", &opt)
	// 	require.NoError(t, err)
	// 	require.Len(t, entries.Rulesets, 5)
	// })
// }

// List returns all rulesets paths because the pathsOnly parameter is set to true.
// // It returns all the entries or just a subset depending on the query string.
// func TestListPaths(t *testing.T) {
// 	t.Parallel()

// 	s, cleanup := newEtcdRulesetService(t)
// 	defer cleanup()

// 	rs := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}

// 	// Root is the basic behaviour without prefix with pathsOnly parameter set to true.
// 	t.Run("Root", func(t *testing.T) {
// 		prefix := "list/paths/root/"

// 		createBoolRuleset(t, s, prefix+"a", rs...)
// 		createBoolRuleset(t, s, prefix+"b", rs...)
// 		createBoolRuleset(t, s, prefix+"a/1", rs...)
// 		createBoolRuleset(t, s, prefix+"c", rs...)
// 		createBoolRuleset(t, s, prefix+"a", rs...)
// 		createBoolRuleset(t, s, prefix+"a/1", rs...)
// 		createBoolRuleset(t, s, prefix+"a/2", rs...)
// 		createBoolRuleset(t, s, prefix+"d", rs...)

// 		paths := []string{prefix + "a", prefix + "a/1", prefix + "a/2", prefix + "b", prefix + "c", prefix + "d"}

// 		opt := api.ListOptions{PathsOnly: true}
// 		entries, err := s.List(context.Background(), prefix+"", &opt)
// 		require.NoError(t, err)
// 		require.Len(t, entries.Rulesets, len(paths))
// 		for i, e := range entries.Rulesets {
// 			require.Equal(t, paths[i], e.Path)
// 			require.Zero(t, e.Rules)
// 			require.Zero(t, e.Version)
// 		}
// 		require.NotEmpty(t, entries.Revision)
// 		require.Zero(t, entries.Continue)
// 	})

// 	// Prefix tests List with a given prefix with pathsOnly parameter set to true.
// 	t.Run("Prefix", func(t *testing.T) {
// 		prefix := "list/paths/prefix/"

// 		createBoolRuleset(t, s, prefix+"x", rs...)
// 		createBoolRuleset(t, s, prefix+"xx", rs...)
// 		createBoolRuleset(t, s, prefix+"x/1", rs...)
// 		createBoolRuleset(t, s, prefix+"xy", rs...)
// 		createBoolRuleset(t, s, prefix+"xy/ab", rs...)
// 		createBoolRuleset(t, s, prefix+"xyz", rs...)

// 		paths := []string{prefix + "xy", prefix + "xy/ab", prefix + "xyz"}

// 		opt := api.ListOptions{PathsOnly: true}
// 		entries, err := s.List(context.Background(), prefix+"xy", &opt)
// 		require.NoError(t, err)
// 		require.Len(t, entries.Rulesets, len(paths))
// 		for i, e := range entries.Rulesets {
// 			require.Equal(t, paths[i], e.Path)
// 			require.Zero(t, e.Rules)
// 			require.Zero(t, e.Version)
// 		}
// 		require.NotEmpty(t, entries.Revision)
// 		require.Zero(t, entries.Continue)
// 	})

// 	// NotFound tests List with a prefix which doesn't exist with pathsOnly parameter set to true.
// 	t.Run("NotFound", func(t *testing.T) {
// 		opt := api.ListOptions{PathsOnly: true}
// 		_, err := s.List(context.Background(), "doesntexist", &opt)
// 		require.Equal(t, err, api.ErrRulesetNotFound)
// 	})

// 	// Paging tests List with pagination with pathsOnly parameter set to true.
// 	t.Run("Paging", func(t *testing.T) {
// 		prefix := "list/paths/paging/"

// 		createBoolRuleset(t, s, prefix+"foo", rs...)
// 		createBoolRuleset(t, s, prefix+"foo/bar", rs...)
// 		createBoolRuleset(t, s, prefix+"foo/bar/baz", rs...)
// 		createBoolRuleset(t, s, prefix+"foo/bar", rs...)
// 		createBoolRuleset(t, s, prefix+"foo/babar", rs...)
// 		createBoolRuleset(t, s, prefix+"foo", rs...)

// 		opt := api.ListOptions{Limit: 2, PathsOnly: true}
// 		entries, err := s.List(context.Background(), prefix+"f", &opt)
// 		require.NoError(t, err)
// 		paths := []string{prefix + "foo", prefix + "foo/babar"}
// 		require.Len(t, entries.Rulesets, len(paths))
// 		for i, e := range entries.Rulesets {
// 			require.Equal(t, paths[i], e.Path)
// 			require.Zero(t, e.Rules)
// 			require.Zero(t, e.Version)
// 		}
// 		require.NotEmpty(t, entries.Revision)
// 		require.NotEmpty(t, entries.Continue)

// 		opt.ContinueToken = entries.Continue
// 		entries, err = s.List(context.Background(), prefix+"f", &opt)
// 		require.NoError(t, err)
// 		paths = []string{prefix + "foo/bar", prefix + "foo/bar/baz"}
// 		require.Len(t, entries.Rulesets, len(paths))
// 		for i, e := range entries.Rulesets {
// 			require.Equal(t, paths[i], e.Path)
// 			require.Zero(t, e.Rules)
// 			require.Zero(t, e.Version)
// 		}
// 		require.NotEmpty(t, entries.Revision)
// 		require.Zero(t, entries.Continue)

// 		opt.ContinueToken = "bad token"
// 		_, err = s.List(context.Background(), prefix+"f", &opt)
// 		require.Equal(t, api.ErrInvalidContinueToken, err)

// 		opt.Limit = -10
// 		opt.ContinueToken = ""
// 		entries, err = s.List(context.Background(), prefix+"f", &opt)
// 		require.NoError(t, err)
// 		paths = []string{prefix + "foo", prefix + "foo/babar", prefix + "foo/bar", prefix + "foo/bar/baz"}
// 		require.Len(t, entries.Rulesets, len(paths))
// 		for i, e := range entries.Rulesets {
// 			require.Equal(t, paths[i], e.Path)
// 			require.Zero(t, e.Rules)
// 			require.Zero(t, e.Version)
// 		}
// 		require.NotEmpty(t, entries.Revision)
// 		require.Zero(t, entries.Continue)
// 	})
// }