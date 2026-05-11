package bep

import (
	"github.com/ndscm/theseed/seed/devprod/rbe/bes/proto/buildeventstreampb"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func searchPatternEvent(
	evs *BuildEventSequence, label string,
) (*buildeventstreampb.BuildEvent, error) {
	results := evs.Search(func(id *buildeventstreampb.BuildEventId) bool {
		for _, p := range id.GetPattern().GetPattern() {
			if p == label {
				return true
			}
		}
		return false
	})
	if len(results) == 0 {
		return nil, seederr.WrapErrorf("no pattern event found for label: %s", label)
	}
	if len(results) > 1 {
		return nil, seederr.WrapErrorf("multiple pattern events found for label: %s", label)
	}
	return results[0], nil
}

func searchTargetConfiguredEvent(
	evs *BuildEventSequence, patternEvent *buildeventstreampb.BuildEvent, label string,
) (*buildeventstreampb.BuildEvent, error) {
	patterns := patternEvent.GetId().GetPattern().GetPattern()
	children := patternEvent.GetChildren()
	if len(patterns) != len(children) {
		return nil, seederr.WrapErrorf("pattern/children count mismatch: %d patterns, %d children", len(patterns), len(children))
	}
	idx := -1
	for i, p := range patterns {
		if p == label {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, seederr.WrapErrorf("label not found in pattern list: %s", label)
	}
	return evs.Get(children[idx])
}

func searchTargetCompletedEvent(
	evs *BuildEventSequence, resolvedLabel string, outputGroup string,
) (*buildeventstreampb.BuildEvent, error) {
	candidates := evs.Search(func(id *buildeventstreampb.BuildEventId) bool {
		tc := id.GetTargetCompleted()
		return tc != nil && tc.GetLabel() == resolvedLabel
	})
	result := (*buildeventstreampb.BuildEvent)(nil)
	for _, event := range candidates {
		completed := event.GetCompleted()
		if completed == nil || !completed.GetSuccess() {
			continue
		}
		for _, og := range completed.GetOutputGroup() {
			if og.GetName() == outputGroup {
				result = event
				break
			}
		}
		if result != nil {
			break
		}
	}
	if result == nil {
		return nil, seederr.WrapErrorf("no successful target completed event found for label %s with output group: %s", resolvedLabel, outputGroup)
	}
	return result, nil
}

func searchNamedSetEvents(
	evs *BuildEventSequence, completedEvent *buildeventstreampb.BuildEvent, outputGroup string,
) ([]*buildeventstreampb.BuildEvent, error) {
	completed := completedEvent.GetCompleted()
	results := []*buildeventstreampb.BuildEvent{}
	for _, og := range completed.GetOutputGroup() {
		if og.GetName() != outputGroup {
			continue
		}
		stack := []*buildeventstreampb.BuildEventId_NamedSetOfFilesId{}
		stack = append(stack, og.GetFileSets()...)
		for len(stack) > 0 {
			setId := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			event, err := evs.Get(&buildeventstreampb.BuildEventId{
				Id: &buildeventstreampb.BuildEventId_NamedSet{
					NamedSet: setId,
				},
			})
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			namedSet := event.GetNamedSetOfFiles()
			if namedSet == nil {
				continue
			}
			results = append(results, event)
			stack = append(stack, namedSet.GetFileSets()...)
		}
	}
	return results, nil
}

// QueryOutput walks the BEP event graph to find output file paths for a
// single target. The label is matched against pattern events that may contain
// multiple patterns, then resolved through:
// pattern → targetConfigured → targetCompleted → outputGroup → namedSet.
func QueryOutput(
	evs *BuildEventSequence, label string, outputGroup string,
) ([]*buildeventstreampb.File, error) {
	if outputGroup == "" {
		outputGroup = "default"
	}
	seedlog.Debugf("Query: label=%s outputGroup=%s", label, outputGroup)
	patternEvent, err := searchPatternEvent(evs, label)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Debugf("Pattern: event=%v", patternEvent)
	configuredEvent, err := searchTargetConfiguredEvent(evs, patternEvent, label)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	canonicalLabel := configuredEvent.GetId().GetTargetConfigured().GetLabel()
	seedlog.Debugf("Configured: canonical=%s", canonicalLabel)
	completedEvent, err := searchTargetCompletedEvent(evs, canonicalLabel, outputGroup)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Debugf("Completed: event=%v", completedEvent)
	namedSetEvents, err := searchNamedSetEvents(evs, completedEvent, outputGroup)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	files := []*buildeventstreampb.File{}
	for _, event := range namedSetEvents {
		seedlog.Debugf("File set: event=%v", event)
		files = append(files, event.GetNamedSetOfFiles().GetFiles()...)
	}
	return files, nil
}
