package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"
	"github.com/ugent-library/muts/store"
)

func init() {
	rootCmd.AddCommand(seedCmd)
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		s, err := store.New(ctx, config.Store.Conn)
		if err != nil {
			return err
		}

		bookID := ulid.Make().String()
		person1ID := ulid.Make().String()
		person2ID := ulid.Make().String()
		chapterID := ulid.Make().String()
		orgIDs := []string{
			ulid.Make().String(),
			ulid.Make().String(),
			ulid.Make().String(),
		}

		err = s.Mutate(ctx,
			store.Mut{
				RecordID: orgIDs[0],
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Organization", nil),
				},
			},
			store.Mut{
				RecordID: orgIDs[1],
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Organization", nil),
					store.AddRel(ulid.Make().String(), "Parent", orgIDs[0], nil),
				},
			},
			store.Mut{
				RecordID: orgIDs[2],
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Organization", nil),
					store.AddRel(ulid.Make().String(), "Parent", orgIDs[1], nil),
				},
			},
			store.Mut{
				RecordID: bookID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Publication.Book", map[string]any{
						"title": map[string]string{"eng": "A treatise on nonsense"},
					}),
				},
			},
			store.Mut{
				RecordID: person1ID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Person", nil),
					store.SetAttr("name", "Mr. Whimsi"),
					store.AddRel(ulid.Make().String(), "Affiliation", orgIDs[0], nil),
				},
			},
			store.Mut{
				RecordID: person2ID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Person", nil),
					store.SetAttr("name", "Mr. Floppy"),
				},
			},
			store.Mut{
				RecordID: chapterID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Publication.Chapter", nil),
					store.SetAttr("title", "Nonsensical introduction"),
					store.AddRel(ulid.Make().String(), "PartOf", bookID, nil),
					store.AddRel(ulid.Make().String(), "Contribution.Author", person1ID, nil),
					store.AddRel(ulid.Make().String(), "Contribution.Author", person2ID, map[string]any{
						"creditRole": "firstAuthor",
					}),
				},
			},
			store.Mut{
				RecordID: chapterID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRel(ulid.Make().String(), "Contribution.FirstAuthor", person2ID, nil),
				},
			},
		)
		if err != nil {
			return err
		}

		// rec, err := s.Rec().WithRelRecs().HasAttr("title").ID(chapterID).One(ctx)
		// if err != nil {
		// 	return err
		// }
		// j, _ := json.MarshalIndent(rec, "", "  ")
		// fmt.Printf("chapter: %s\n", j)

		// err = s.Rec().WithRels().Kind("Person").Each(ctx, func(rec *store.Rec) bool {
		// 	j, _ := json.Marshal(rec)
		// 	fmt.Printf("person: %s\n", j)
		// 	return true
		// })
		// if err != nil {
		// 	return err
		// }

		// err = s.Rec().KindMatches("Publication.*").Each(ctx, func(rec *store.Rec) bool {
		// 	j, _ := json.Marshal(rec)
		// 	fmt.Printf("publication: %s\n", j)
		// 	return true
		// })
		// if err != nil {
		// 	return err
		// }

		// err = s.Rec().RelKind("PartOf").Each(ctx, func(rec *store.Rec) bool {
		// 	j, _ := json.Marshal(rec)
		// 	fmt.Printf("record that is PartOf something: %s\n", j)
		// 	return true
		// })
		// if err != nil {
		// 	return err
		// }

		// select * from muts_select('{"id": "01HSP4V7VCFSBRF7MECEH4KDRX"}');
		// select * from muts_select('{"id": "01HSP4V7VCFSBRF7MECEH4KDRX", "follow":"PartOf|Contribution.*"}');
		// select * from muts_select('{"id_in": ["01HSP4V7VCFSBRF7MECEH4KDRX", "01HSP4V7VCFSBRF7MECNWHJ1A5"]}');
		// select * from muts_select('{"kind": "Publication.Chapter"}');
		// select * from muts_select('{"kind": "Publication.*"}');
		// select * from muts_select('{"attr": "$.title"}');
		recs, err := s.Many(ctx, store.Query{Limit: 10})
		if err != nil {
			return err
		}
		j, _ := json.Marshal(recs)
		fmt.Printf("many: %s\n", j)

		return nil
	},
}
