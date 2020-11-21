// Copyright © 2020 David Arnold <dar@xoe.solutions>
// SPDX-License-Identifier: MIT

package directive

import (
	"bytes"
	"fmt"
	. "github.com/dave/jennifer/jen"
	// "log"
	"regexp"
	"strings"
)

var cmdGenCommand string = "ddd-gen app command"

type genTyp string

const (
	AddTyp genTyp = "Add"
	RemTyp genTyp = "Remove"
	UpdTyp genTyp = "Update"
)

type QualId struct{ Id, Qual string }

type NamedQualId struct {
	Name string
	QualId
}

type ParsedConfig struct {
	// Aggregate Entity
	AggEntityStruct QualId
	// Application defined interfaces
	PoliceableInterface   QualId
	IdentifiableInterface QualId
	RepositoryInterface   QualId
	PolicerInterface      QualId
	// Iface method return types
	IdentifierTyp     QualId
	UserTyp           QualId
	ElevationTokenTyp QualId
	// Error constructor
	AuthorizationErrorNew  QualId
	IdentificationErrorNew QualId
	RepositoryErrorNew     QualId
	DomainErrorNew         QualId
}

// Regeneratables ...

func addCommandErrors(f *File, DoSomething string, authErr, idErr, repoErr, domErr *QualId) {
	f.Null().Var().DefsFunc(func(g *Group) {
		if authErr != nil {
			g.Commentf("ErrNotAuthorizedTo%s signals that the caller is not authorized to perform %s", DoSomething, DoSomething)
			g.Id("ErrNotAuthorizedTo"+DoSomething).Op("=").Qual(
				authErr.Qual,
				authErr.Id,
			).Call(
				Lit("ErrNotAuthorizedTo" + DoSomething),
			)
		}
		if idErr != nil {
			g.Commentf("Err%sNotIdentifiable signals that %s's command object was not identifiable", DoSomething, DoSomething)
			g.Id("Err"+DoSomething+"NotIdentifiable").Op("=").Qual(
				idErr.Qual,
				idErr.Id,
			).Call(
				Lit("Err" + DoSomething + "NotIdentifiable"),
			)
		}
		if repoErr != nil {
			g.Commentf("Err%sFailedInRepository signals that %s failed in the repository layer", DoSomething, DoSomething)
			g.Id("Err"+DoSomething+"FailedInRepository").Op("=").Qual(
				repoErr.Qual,
				repoErr.Id,
			).Call(
				Lit("Err" + DoSomething + "FailedInRepository"),
			)
		}
		if domErr != nil {
			g.Commentf("Err%sFailedInDomain signals that %s failed in the domain layer", DoSomething, DoSomething)
			g.Id("Err"+DoSomething+"FailedInDomain").Op("=").Qual(
				domErr.Qual,
				domErr.Id,
			).Call(
				Lit("Err" + DoSomething + "FailedInDomain"),
			)
		}
	})
}

func addCommandHandlerType(f *File, DoSomething string, adapters []NamedQualId) {
	f.Commentf("%sHandler knows how to perform %s", DoSomething, DoSomething)
	f.Null().Type().Id(
		DoSomething + "Handler",
	).StructFunc(func(g *Group) {
		for _, a := range adapters {
			g.Id(a.Name).Qual(a.Qual, a.Id)
		}
	})
}
func addCommandHandlerConstructor(f *File, DoSomething string, adapters []NamedQualId) {
	f.Commentf("New%sHandler returns %sHandler", DoSomething, DoSomething)
	f.Func().Id(
		"New" + DoSomething + "Handler",
	).ParamsFunc(func(g *Group) {
		for _, a := range adapters {
			g.Id(a.Name).Qual(a.Qual, a.Id)
		}
	}).Params(
		Op("*").Id(DoSomething + "Handler"),
	).BlockFunc(func(g *Group) {
		for _, a := range adapters {
			g.If(
				Qual(
					"reflect", "ValueOf",
				).Call(
					Id(a.Name),
				).Dot(
					"IsZero",
				).Call(),
			).Block(
				Id("panic").Call(Lit("no '" + a.Name + "' provided!")),
			)
		}
		g.Return().Op("&").Id(
			DoSomething + "Handler",
		).ValuesFunc(func(g *Group) {
			for _, a := range adapters {
				g.Id(a.Name).Op(":").Id(a.Name)
			}
		})
	})
}
func addCommandFuncHandle(f *File,
	DoSomething string,
	withPolicy bool,
	aggEntity QualId,
	genTyp genTyp,
	extraAdapters []NamedQualId) {
	entityShort := cmdShortForm(aggEntity.Id)
	f.Commentf("Handle generically performs %s", DoSomething)
	f.Func().Params(
		Id("h").Id(DoSomething+"Handler"),
	).Id(
		"Handle",
	).Params(
		Id("ctx").Qual("context", "Context"),
		Id(cmdShortForm(DoSomething)).Id(DoSomething),
	).Parens(
		List(
			Id("error"),
		),
	).BlockFunc(func(g *Group) {
		g.If(
			Op("!").Id(cmdShortForm(DoSomething)).Dot("IsIdentifiable").Call(),
		).Block(
			Return().Id("Err" + DoSomething + "NotIdentifiable"),
		)
		g.List(
			Id("err"),
		).Op(":=").Id("h").Dot("agg").Dot(
			string(genTyp),
		).Call(
			Id("ctx"),
			Id(cmdShortForm(DoSomething)),
			Func().Params(
				Id(entityShort).Op("*").Qual(aggEntity.Qual, aggEntity.Id),
			).Parens(
				List(
					Id("error"),
				),
			).BlockFunc(func(g *Group) {
				if withPolicy {
					g.List(
						Id("data"),
						Id("err"),
					).Op(":=").Qual("encoding/json", "Marshal").Call(Id(entityShort))
					g.If(
						Id("err").Op("!=").Id("nil"),
					).Block(
						Panic(
							Id("err"),
						).Comment(
							"invariant violation: the domain shall always be consistent!",
						),
					)
					g.If(
						Id("ok").Op(":=").Id("h").Dot("pol").Dot(
							"Can",
						).Call(
							Id("ctx"),
							Id(cmdShortForm(DoSomething)),
							Lit(DoSomething),
							Id("data"),
						),
						Op("!").Id("ok"),
					).Block(
						Return().Id("ErrNotAuthorizedTo" + DoSomething),
					)
				}
				g.If(
					Id("err").Op(":=").Id(cmdShortForm(DoSomething)).Dot(
						"handle",
					).CallFunc(func(g *Group) {
						g.Id("ctx")
						g.Id(entityShort)
						for _, a := range extraAdapters {
							g.Op("&").Id("h").Dot(a.Name)
						}
					}),
					Id("err").Op("!=").Id("nil"),
				).Block(
					Return().Qual("github.com/hashicorp/errwrap", "Wrap").Call(
						Id("Err"+DoSomething+"FailedInDomain"),
						Id("err"),
					),
				)
				g.Return().Id("nil")
			}),
		)
		g.If(
			Id("err").Op("!=").Id("nil"),
		).BlockFunc(func(g *Group) {
			if withPolicy {
				g.If(
					Qual("github.com/hashicorp/errwrap", "Contains").Call(
						Id("err"),
						Lit("ErrNotAuthorizedTo"+DoSomething),
					),
				).Block(
					Return().Id("err"),
				)
			}
			g.If(
				Qual("github.com/hashicorp/errwrap", "Contains").Call(
					Id("err"),
					Lit("Err"+DoSomething+"FailedInDomain"),
				),
			).Block(
				Return().Id("err"),
			)
			g.Return(
				Qual("github.com/hashicorp/errwrap", "Wrap").Call(
					Lit("Err"+DoSomething+"FailedInRepository"),
					Id("err"),
				),
			)
		})
		g.Return(
			Id("nil"),
		)
	})
}

func addCommandFuncHandleNew(f *File,
	DoSomething string,
	withPolicy,
	addWithIdentifiable bool,
	aggEntity,
	identifierTyp QualId,
	genTyp genTyp,
	extraAdapters []NamedQualId) {
	entityShort := cmdShortForm(aggEntity.Id)
	needReturnIdentifer := (!addWithIdentifiable)
	f.Commentf("Handle generically performs %s", DoSomething)
	f.Func().Params(
		Id("h").Id(DoSomething+"Handler"),
	).Id(
		"Handle",
	).Params(
		Id("ctx").Qual("context", "Context"),
		Id(cmdShortForm(DoSomething)).Id(DoSomething),
	).Parens(
		ListFunc(func(g *Group) {
			if needReturnIdentifer {
				g.Qual(identifierTyp.Qual, identifierTyp.Id)
			}
			g.Id("error")
		}),
	).BlockFunc(func(g *Group) {
		if addWithIdentifiable {
			g.If(
				Op("!").Id(cmdShortForm(DoSomething)).Dot("IsIdentifiable").Call(),
			).Block(
				Return().Id("Err" + DoSomething + "NotIdentifiable"),
			)
		}
		g.ListFunc(func(g *Group) {
			if needReturnIdentifer {
				g.Id("identifier")
			} else {
				g.Id("_")
			}
			g.Id("err")
		}).Op(":=").Id("h").Dot("agg").Dot(
			string(genTyp),
		).Call(
			Id("ctx"),
			Func().Params().Parens(
				List(
					Id(entityShort).Op("*").Qual(aggEntity.Qual, aggEntity.Id),
					Id("err").Id("error"),
				),
			).BlockFunc(func(g *Group) {
				g.If(
					List(
						Id("err"),
					).Op(":=").Id(cmdShortForm(DoSomething)).Dot(
						"handle",
					).CallFunc(func(g *Group) {
						g.Id("ctx")
						g.Id(entityShort)
						for _, a := range extraAdapters {
							g.Op("&").Id("h").Dot(a.Name)
						}
					}),
					Id("err").Op("!=").Id("nil"),
				).Block(
					Return().Qual("github.com/hashicorp/errwrap", "Wrap").Call(
						Id("Err"+DoSomething+"FailedInDomain"),
						Id("err"),
					),
				)
				if withPolicy {
					g.List(
						Id("data"),
						Id("err"),
					).Op(":=").Qual("encoding/json", "Marshal").Call(Id(entityShort))
					g.If(
						Id("err").Op("!=").Id("nil"),
					).Block(
						Panic(
							Id("err"),
						).Comment(
							"invariant violation: the domain shall always be consistent!",
						),
					)
					g.If(
						Id("ok").Op(":=").Id("h").Dot("pol").Dot(
							"Can",
						).Call(
							Id("ctx"),
							Id(cmdShortForm(DoSomething)),
							Lit(DoSomething),
							Id("data"),
						),
						Op("!").Id("ok"),
					).Block(
						Return(
							Id("nil"), Id("ErrNotAuthorizedTo"+DoSomething),
						),
					)
				}
				g.Return(
					Id(entityShort), Id("nil"),
				)
			}),
		)
		g.If(
			Id("err").Op("!=").Id("nil"),
		).BlockFunc(func(g *Group) {
			if withPolicy {
				g.If(
					Qual("github.com/hashicorp/errwrap", "Contains").Call(
						Id("err"),
						Lit("ErrNotAuthorizedTo"+DoSomething),
					),
				).Block(
					ReturnFunc(func(g *Group) {
						if needReturnIdentifer {
							g.Id("identifier")
							g.Id("err")
						} else {
							g.Id("err")
						}
					}),
				)
			}
			g.If(
				Qual("github.com/hashicorp/errwrap", "Contains").Call(
					Id("err"),
					Lit("Err"+DoSomething+"FailedInDomain"),
				),
			).Block(
				ReturnFunc(func(g *Group) {
					if needReturnIdentifer {
						g.Id("identifier")
						g.Id("err")
					} else {
						g.Id("err")
					}
				}),
			)
			g.ReturnFunc(func(g *Group) {
				if needReturnIdentifer {
					g.Id("identifier")
					g.Qual("github.com/hashicorp/errwrap", "Wrap").Call(
						Lit("Err"+DoSomething+"FailedInRepository"),
						Id("err"),
					)
				} else {
					g.Qual("github.com/hashicorp/errwrap", "Wrap").Call(
						Lit("Err"+DoSomething+"FailedInRepository"),
						Id("err"),
					)
				}
			})
		})
		g.ReturnFunc(func(g *Group) {
			if needReturnIdentifer {
				g.Id("identifier")
				g.Id("nil")
			} else {
				g.Id("nil")
			}
		})
	})
}

func addCommandIsIdentifiable(f *File, DoSomething string) {
	f.Commentf("IsIdentifiable answers the question wether the %s's object is identifiable", DoSomething)
	f.Null().Func().Params(
		Id(cmdShortForm(DoSomething)).Id(DoSomething),
	).Id(
		"IsIdentifiable",
	).Params().Id("bool").Block(
		If(
			Qual("reflect", "ValueOf").Call(
				Id(cmdShortForm(DoSomething)).Dot("Identifier").Call(),
			).Dot("IsZero").Call(),
		).Block(
			Return().Id("true"),
		),
		Return().Id("false"),
	)
}

// Stubs ...

func addCommandTypeStub(f *File, DoSomething string, identifierTyp, userTyp, elevationTokenTyp QualId) {
	f.Commentf("%s represents a %s command", DoSomething, DoSomething)
	f.Null().Type().Id(
		DoSomething,
	).Struct(
		Id("uuid").Qual(identifierTyp.Qual, identifierTyp.Id),
		Id("userId").Qual(userTyp.Qual, userTyp.Id),
		Id("elevationToken").Qual(elevationTokenTyp.Qual, elevationTokenTyp.Id),
		Line(),
		Comment("TODO: design command event/message fields (evtl. use protobuf + protoc-gen-go)"),
	)
}

func addCommandHandleStub(f *File, DoSomething string, aggEntity QualId, extraAdapters []NamedQualId) {
	entityShort := cmdShortForm(aggEntity.Id)
	f.Commentf("handle specifically performs %s", DoSomething)
	f.Null().Func().Params(
		Id(cmdShortForm(DoSomething)).Op("*").Id(DoSomething),
	).Id(
		"handle",
	).ParamsFunc(func(g *Group) {
		g.Id("ctx").Qual("context", "Context")
		g.Id(entityShort).Op("*").Qual(aggEntity.Qual, aggEntity.Id)
		for _, a := range extraAdapters {
			g.Id(a.Name).Op("*").Qual(a.Qual, a.Id)
		}
	}).Add(
		Id("error"),
	).Block(
		Comment("TODO: implement app logic"),
		Return().Id("nil"),
	)
}

func addCommandIdentifierStub(f *File, DoSomething string, identifierTyp QualId) {
	f.Commentf("Identifier returns the identifier of the object on which to perform %s", DoSomething)
	f.Null().Func().Params(
		Id(cmdShortForm(DoSomething)).Id(DoSomething),
	).Id(
		"Identifier",
	).Params().Add(
		Qual(identifierTyp.Qual, identifierTyp.Id),
	).Block(
		Return().Id(cmdShortForm(DoSomething)).Dot("uuid"),
	)
}

func addCommandUserStub(f *File, DoSomething string, userTyp QualId) {
	f.Commentf("User returns the identifier of the caller")
	f.Null().Func().Params(
		Id(cmdShortForm(DoSomething)).Id(DoSomething),
	).Id(
		"User",
	).Params().Add(
		Qual(userTyp.Qual, userTyp.Id),
	).Block(
		Return().Id(cmdShortForm(DoSomething)).Dot("userId"),
	)
}

func addCommandElevationTokenStub(f *File, DoSomething string, elevationTokenTyp QualId) {
	f.Commentf("ElevationToken returns an elevation token in posession of the caller")
	f.Null().Func().Params(
		Id(cmdShortForm(DoSomething)).Id(DoSomething),
	).Id(
		"ElevationToken",
	).Params().Add(
		Qual(elevationTokenTyp.Qual, elevationTokenTyp.Id),
	).Block(
		Return().Id(cmdShortForm(DoSomething)).Dot("elevationToken"),
	)
}

func addCommandInterfaceAssertionIdentifiable(f *File, DoSomething string, identifiable QualId) {
	f.Commentf("Assert that %s implements Identifiable interface!", DoSomething)
	f.Var().Id("_").Qual(identifiable.Qual, identifiable.Id).Op("=").Parens(Op("*").Id(DoSomething)).Call(Id("nil"))
}

func addCommandInterfaceAssertionPoliceable(f *File, DoSomething string, policeable QualId) {
	f.Commentf("Assert that %s implements Policeable interface!", DoSomething)
	f.Var().Id("_").Qual(policeable.Qual, policeable.Id).Op("=").Parens(Op("*").Id(DoSomething)).Call(Id("nil"))
}

// Composers ...

func GenCommand(cmd, topic string, withPolicy, addWithIdentifiable bool, allAdapters, extraAdapters []NamedQualId, genTyp genTyp, conf ParsedConfig) *File {
	ret := NewFile("command")
	ret.HeaderComment(fmt.Sprintf("Code generated by '%s': DO NOT EDIT.", cmdGenCommand))
	ret.Line()
	ret.Commentf("Topic: %s", topic)
	ret.Line()
	if genTyp != AddTyp || addWithIdentifiable {
		if withPolicy {
			addCommandErrors(ret, cmd,
				&conf.AuthorizationErrorNew,
				&conf.IdentificationErrorNew,
				&conf.RepositoryErrorNew,
				&conf.DomainErrorNew)
		} else {
			addCommandErrors(ret, cmd,
				nil,
				&conf.IdentificationErrorNew,
				&conf.RepositoryErrorNew,
				&conf.DomainErrorNew)
		}
	} else {
		if withPolicy {
			addCommandErrors(ret, cmd,
				&conf.AuthorizationErrorNew,
				nil,
				&conf.RepositoryErrorNew,
				&conf.DomainErrorNew)
		} else {
			addCommandErrors(ret, cmd,
				nil,
				nil,
				&conf.RepositoryErrorNew,
				&conf.DomainErrorNew)
		}
	}
	addCommandHandlerType(ret, cmd, allAdapters)
	addCommandHandlerConstructor(ret, cmd, allAdapters)
	if genTyp != AddTyp {
		addCommandFuncHandle(ret, cmd, withPolicy, conf.AggEntityStruct, genTyp, extraAdapters)
	} else {
		addCommandFuncHandleNew(ret, cmd, withPolicy, addWithIdentifiable, conf.AggEntityStruct, conf.IdentifierTyp, genTyp, extraAdapters)
	}
	return ret
}

func StubCommand(cmd, topic string, withPolicy, withCommandStub, addWithIdentifiable bool, extraAdapters []NamedQualId, genTyp genTyp, conf ParsedConfig) *File {
	ret := NewFile("command")
	ret.HeaderComment(fmt.Sprintf("Code generated by '%s': THESE ARE STUBS, PLEASE EDIT.", cmdGenCommand))
	ret.Line()
	ret.Commentf("/*\n\t=== Topic: %s ===\n*/", topic)
	ret.Line()
	if withCommandStub {
		addCommandTypeStub(ret, cmd, conf.IdentifierTyp, conf.UserTyp, conf.ElevationTokenTyp)
	}
	addCommandHandleStub(ret, cmd, conf.AggEntityStruct, extraAdapters)
	ret.Line()
	ret.Comment("/*\n\t=> Identifiable interface implementation ...\n*/")
	ret.Line()
	if genTyp != AddTyp || addWithIdentifiable {
		addCommandIdentifierStub(ret, cmd, conf.IdentifierTyp)
		addCommandInterfaceAssertionIdentifiable(ret, cmd, conf.IdentifiableInterface)
	}
	if withPolicy {
		ret.Comment("/*\n\t=> Policeable interface implementation ...\n*/")
		ret.Line()
		addCommandUserStub(ret, cmd, conf.UserTyp)
		addCommandElevationTokenStub(ret, cmd, conf.ElevationTokenTyp)
		addCommandInterfaceAssertionPoliceable(ret, cmd, conf.PoliceableInterface)
	}
	return ret
}

func GenDoc(docFile string) *File {
	ret := NewFile("command")
	ret.PackageComment("Package command implements application layer commands")
	return ret
}

// Utils ...

func cmdShortForm(s string) string {
	re := regexp.MustCompile(`[A-Z]`)
	var b bytes.Buffer
	for _, el := range re.FindAllString(s, -1) {
		b.WriteString(strings.ToLower(el))
	}
	return b.String()
}

func splitQual(s string) (string, string) {
	imp := s[:strings.LastIndex(s, ".")]
	id := s[strings.LastIndex(s, ".")+1:]
	return imp, id
}
