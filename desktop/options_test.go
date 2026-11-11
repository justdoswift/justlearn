package main

import (
	"strings"
	"testing"
)

func legacyQuizSample() string {
	return "orders 表里一个用户可有多条订单。要回答「上月有多少复购用户」，哪段 SQL 最终数的是人？\n\n" +
		"• A. `select count(*) from orders where created_at >= date '2024-05-01' group by user_id`\n" +
		"• B. `select count(*) from (select user_id from orders group by user_id having count(*) >= 2) t`\n" +
		"• C. `select count(user_id) from orders where order_id >= 2`\n\n" +
		"只要字母即可。"
}

func TestSplitLegacyOptions(t *testing.T) {
	before, options, after, ok := splitLegacyOptions(legacyQuizSample())
	if !ok {
		t.Fatal("legacy choice list was not recognized")
	}
	if !strings.Contains(before, "哪段 SQL") {
		t.Fatalf("before=%q", before)
	}
	if !strings.Contains(after, "只要字母即可") {
		t.Fatalf("after=%q", after)
	}
	if len(options) != 3 {
		t.Fatalf("options=%d, want 3", len(options))
	}
	if options[0].Letter != "A" || options[2].Letter != "C" {
		t.Fatalf("letters=%q/%q", options[0].Letter, options[2].Letter)
	}
	if !options[1].Code || !strings.Contains(options[1].Text, "select count") {
		t.Fatalf("second option was not recognized as code: %#v", options[1])
	}
}

func TestSplitLegacyOptionsSecondGroup(t *testing.T) {
	first := "第一题？\n- A. `select 1`\n- B. `select 2`"
	second := "第二题？\n- A. `select 3`\n- B. `select 4`"
	_, options, after, ok := splitLegacyOptions(first + "\n\n" + second)
	if !ok || len(options) != 2 || !strings.Contains(after, "第二题") {
		t.Fatalf("first group: ok=%v options=%d after=%q", ok, len(options), after)
	}
	_, options, _, ok = splitLegacyOptions(after)
	if !ok || len(options) != 2 || options[0].Text != "select 3" {
		t.Fatalf("second group not parsed: ok=%v options=%#v", ok, options)
	}
}

func TestSplitLegacyOptionsLeavesSingleLetterProseAlone(t *testing.T) {
	if _, _, _, ok := splitLegacyOptions("A. 这是一段普通说明，没有第二个选项。"); ok {
		t.Fatal("a single A. line should stay Markdown")
	}
}
