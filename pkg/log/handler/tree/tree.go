package tree

import (
	"log/slog"
)

type Value struct {
	Name      string
	SlogValue slog.Value // must not be slog.KindGroup
	Group     *Group
}

func (value *Value) LogValue() slog.Value {
	if value.Group != nil {
		return slog.GroupValue(value.Group.Render()...)
	}

	return value.SlogValue
}

func (value *Value) clone() *Value {
	newValue := &Value{Name: value.Name}

	if value.Group != nil {
		newValue.Group = value.Group.Clone()
		return newValue
	}

	newValue.SlogValue = value.SlogValue

	return newValue
}

type Group struct {
	Values []*Value
}

func (group *Group) LogValue() slog.Value {
	return slog.GroupValue(group.Render()...)
}

func (group *Group) Clone() *Group {
	newGroup := &Group{Values: make([]*Value, len(group.Values))}

	for i, value := range group.Values {
		newGroup.Values[i] = value.clone()
	}

	return newGroup
}

func (group *Group) Render() []slog.Attr {
	attrs := make([]slog.Attr, len(group.Values))

	for i, value := range group.Values {
		attrs[i] = slog.Attr{Key: value.Name, Value: value.LogValue()}
	}

	return attrs
}

func (group *Group) Merge(stack []string, attrs ...slog.Attr) {
	for i := range attrs {
		merge(group, stack, attrs[i])
	}
}

// merge applies attr on stack to g, iterating recursively and merging where necessary.
func merge(in *Group, stack []string, attr slog.Attr) {
	l := last(in, stack)

	for i := range l.Values {
		if l.Values[i].Name == attr.Key {
			if attr.Value.Kind() != slog.KindGroup {
				l.Values[i].SlogValue = attr.Value
				l.Values[i].Group = nil
				return
			}

			if l.Values[i].Group == nil {
				l.Values[i].Group = new(Group)
			}

			l.Values[i].Group.Merge(nil, attr.Value.Group()...)
			return
		}
	}

	if attr.Value.Kind() != slog.KindGroup {
		l.Values = append(l.Values, &Value{Name: attr.Key, SlogValue: attr.Value})
	} else {
		newGroup := new(Group)
		newGroup.Merge(nil, attr.Value.Group()...)
		l.Values = append(l.Values, &Value{Name: attr.Key, Group: newGroup})
	}
}

func last(in *Group, stack []string) *Group {
	if len(stack) == 0 {
		return in
	}

	for _, v := range in.Values {
		if v.Name == stack[0] {
			return last(v.Group, stack[1:])
		}
	}

	panic("invalid or misconfigured stack")
}
