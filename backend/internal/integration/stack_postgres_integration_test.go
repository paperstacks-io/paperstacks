package integration

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
	stackPostgres "github.com/paperstacks.io/paperstacks/internal/stack/repository/postgres"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

// This test uses an initialized test database, with app_rw credentials. It only
// removes its own uniquely named records and never resets an existing database.
func TestStackPostgresRepository(t *testing.T) {
	dsn := os.Getenv("STACK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set STACK_TEST_DATABASE_URL to run the PostgreSQL repository test")
	}
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.MaxConns = 1
	poolConfig.MinConns = 0
	poolConfig.MinIdleConns = 0
	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	repo := stackPostgres.NewRepository(db)
	suffix := uuid.New().String()
	owner := userDomain.NewUser("stack-test-"+suffix, suffix+"@example.invalid")
	otherOwner := userDomain.NewUser("other-"+suffix, "other-"+suffix+"@example.invalid")
	paper := paperDomain.Paper{UUID: uuid.New().String(), Title: "Repository integration paper"}
	stack := *stackDomain.NewStack("Alpha "+suffix, owner)
	stack.IsPublic = true
	private := *stackDomain.NewStack("Beta "+suffix, owner)
	other := *stackDomain.NewStack("Gamma "+suffix, otherOwner)
	failedID := uuid.New().String()
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, id := range []string{stack.UUID, private.UUID, other.UUID, failedID} {
			if _, err := db.Exec(cleanup, `DELETE FROM public.stack WHERE uuid=$1`, id); err != nil {
				t.Error(err)
			}
		}
		if _, err := db.Exec(cleanup, `DELETE FROM public.paper WHERE uuid=$1`, paper.UUID); err != nil {
			t.Error(err)
		}
		for _, id := range []string{owner.ExternalID, otherOwner.ExternalID} {
			if _, err := db.Exec(cleanup, `DELETE FROM public.app_user WHERE external_id=$1`, id); err != nil {
				t.Error(err)
			}
		}
	})
	for _, user := range []userDomain.User{owner, otherOwner} {
		if _, err := db.Exec(ctx, `INSERT INTO public.app_user (external_id,email) VALUES ($1,$2)`, user.ExternalID, user.Email); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(ctx, `INSERT INTO public.paper (uuid,title) VALUES ($1,$2)`, paper.UUID, paper.Title); err != nil {
		t.Fatal(err)
	}
	baseline, err := repo.CountPublic(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []stackDomain.Stack{stack, private, other} {
		if err := repo.Create(ctx, s); err != nil {
			t.Fatal(err)
		}
	}
	count, err := repo.CountPublic(ctx)
	if err != nil || count != baseline+1 {
		t.Fatalf("public count: %d, %v", count, err)
	}
	duplicate := stack
	duplicate.UUID = uuid.New().String()
	duplicate.Name = "ALPHA " + suffix
	if err := repo.Create(ctx, duplicate); !errors.Is(err, stackDomain.ErrStackAlreadyExists) {
		t.Fatalf("duplicate name: %v", err)
	}
	for range 2 {
		if err := repo.AddPaper(ctx, stack.UUID, paper); err != nil {
			t.Fatal(err)
		}
	}
	loaded, err := repo.GetByUUID(ctx, stack.UUID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Owner.Email != owner.Email || len(loaded.Papers) != 1 || loaded.Papers[0].Title != paper.Title {
		t.Fatalf("loaded stack: %+v", loaded)
	}
	t.Run("concurrent reads share one connection", func(t *testing.T) {
		var workers sync.WaitGroup
		for range 6 {
			workers.Add(1)
			go func() {
				defer workers.Done()
				got, err := repo.GetByUUID(ctx, stack.UUID)
				if err != nil || len(got.Papers) != 1 {
					t.Errorf("concurrent read: %+v, %v", got, err)
				}
			}()
		}
		workers.Wait()
	})
	stacks, err := repo.List(ctx, owner.ExternalID)
	if err != nil || len(stacks) != 2 {
		t.Fatalf("owner list: %+v, %v", stacks, err)
	}
	stacks, err = repo.ListPublic(ctx, owner.ExternalID)
	if err != nil || len(stacks) != 1 || stacks[0].UUID != stack.UUID {
		t.Fatalf("public owner list: %+v, %v", stacks, err)
	}
	stats, err := repo.StatsByOwner(ctx, owner.ExternalID)
	if err != nil || stats != (stackDomain.Stats{TotalStacks: 2, PublicStacks: 1, TotalPapers: 1}) {
		t.Fatalf("stats: %+v, %v", stats, err)
	}
	result, err := repo.SearchByOwner(ctx, owner.ExternalID, stackDomain.SearchOptions{Query: suffix, SortBy: "name", Page: 1, PageSize: 1})
	if err != nil || result.Total != 2 || len(result.Items) != 1 || result.Items[0].UUID != stack.UUID || !result.HasNext {
		t.Fatalf("search: %+v, %v", result, err)
	}
	result, err = repo.SearchPublic(ctx, stackDomain.SearchOptions{Query: suffix, SortBy: "updated_at", Page: 1, PageSize: 10})
	if err != nil || result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("public search: %+v, %v", result, err)
	}
	wrongOwner := loaded
	wrongOwner.Owner = otherOwner
	if _, err := repo.Update(ctx, wrongOwner); !errors.Is(err, stackDomain.ErrStackNotFound) {
		t.Fatalf("wrong owner: %v", err)
	}
	conflicting := loaded
	conflicting.Name = private.Name
	if _, err := repo.Update(ctx, conflicting); !errors.Is(err, stackDomain.ErrStackAlreadyExists) {
		t.Fatalf("update with duplicate name: %v", err)
	}
	loaded.Name = "Renamed " + suffix
	loaded.UpdatedAt = time.Now()
	updated, err := repo.Update(ctx, loaded)
	if err != nil || !updated.CreatedAt.Equal(loaded.CreatedAt) {
		t.Fatalf("update: %+v, %v", updated, err)
	}
	// A failed association replacement must roll back the name and old links.
	invalid := updated
	invalid.Name = "Rollback " + suffix
	invalid.Papers = []paperDomain.Paper{{UUID: uuid.New().String()}}
	if _, err := repo.Update(ctx, invalid); err == nil {
		t.Fatal("expected missing paper failure")
	}
	loaded, err = repo.GetByUUID(ctx, stack.UUID)
	if err != nil || loaded.Name != updated.Name || len(loaded.Papers) != 1 {
		t.Fatalf("rollback: %+v, %v", loaded, err)
	}
	invalid.UUID = failedID
	if err := repo.Create(ctx, invalid); err == nil {
		t.Fatal("expected failed create")
	}
	if _, err := repo.GetByUUID(ctx, failedID); !errors.Is(err, stackDomain.ErrStackNotFound) {
		t.Fatalf("create rollback: %v", err)
	}
	if err := repo.AddPaper(ctx, stack.UUID, paperDomain.Paper{UUID: uuid.New().String()}); !errors.Is(err, paperDomain.ErrPaperNotFound) {
		t.Fatalf("missing paper: %v", err)
	}
	for range 2 {
		if err := repo.RemovePaper(ctx, stack.UUID, paper.UUID); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.AddPaper(ctx, stack.UUID, paper); err != nil {
		t.Fatal(err)
	}
	// A soft-deleted paper stays visible as a reference, but is not counted as active.
	if _, err := db.Exec(ctx, `UPDATE public.paper SET deleted_at=statement_timestamp() WHERE uuid=$1`, paper.UUID); err != nil {
		t.Fatal(err)
	}
	loaded, err = repo.GetByUUID(ctx, stack.UUID)
	if err != nil || len(loaded.Papers) != 0 || len(loaded.DeletedPapers) != 1 || loaded.DeletedPapers[0].UUID != paper.UUID || loaded.DeletedPapers[0].Title != paper.Title {
		t.Fatalf("deleted paper reference missing: %+v, %v", loaded, err)
	}
	// A normal stack edit must not discard references omitted from the active papers list.
	loaded.DeletedPapers = nil
	edited, err := repo.Update(ctx, loaded)
	if err != nil || len(edited.DeletedPapers) != 1 || edited.DeletedPapers[0].UUID != paper.UUID {
		t.Fatalf("stack edit lost deleted paper reference: %+v, %v", edited, err)
	}
	stats, err = repo.StatsByOwner(ctx, owner.ExternalID)
	if err != nil || stats.TotalPapers != 0 {
		t.Fatalf("deleted paper counted: %+v, %v", stats, err)
	}
	if err := repo.AddPaper(ctx, private.UUID, paper); !errors.Is(err, paperDomain.ErrPaperNotFound) {
		t.Fatalf("add deleted paper: %v", err)
	}
	if err := repo.Delete(ctx, stack.UUID); err != nil {
		t.Fatal(err)
	}
	var deletedAt, repeatedDeletedAt time.Time
	if err := db.QueryRow(ctx, `SELECT deleted_at FROM public.stack WHERE uuid=$1`, stack.UUID).Scan(&deletedAt); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, stack.UUID); !errors.Is(err, stackDomain.ErrStackNotFound) {
		t.Fatalf("missing stack: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT deleted_at FROM public.stack WHERE uuid=$1`, stack.UUID).Scan(&repeatedDeletedAt); err != nil || !deletedAt.Equal(repeatedDeletedAt) {
		t.Fatalf("repeated delete changed retention timestamp: %v", err)
	}
	if _, err := repo.GetByUUID(ctx, stack.UUID); !errors.Is(err, stackDomain.ErrStackNotFound) {
		t.Fatalf("get deleted stack: %v", err)
	}
	if _, err := repo.Update(ctx, updated); !errors.Is(err, stackDomain.ErrStackNotFound) {
		t.Fatalf("update deleted stack: %v", err)
	}
	if err := repo.AddPaper(ctx, stack.UUID, paper); !errors.Is(err, stackDomain.ErrStackNotFound) {
		t.Fatalf("add to deleted stack: %v", err)
	}
	if err := repo.RemovePaper(ctx, stack.UUID, paper.UUID); !errors.Is(err, stackDomain.ErrStackNotFound) {
		t.Fatalf("remove from deleted stack: %v", err)
	}
	stacks, err = repo.List(ctx, owner.ExternalID)
	if err != nil || len(stacks) != 1 || stacks[0].UUID != private.UUID {
		t.Fatalf("list deleted stack: %+v, %v", stacks, err)
	}
	stacks, err = repo.ListPublic(ctx, owner.ExternalID)
	if err != nil || len(stacks) != 0 {
		t.Fatalf("public list deleted stack: %+v, %v", stacks, err)
	}
	result, err = repo.SearchPublic(ctx, stackDomain.SearchOptions{Query: suffix, SortBy: "name"})
	if err != nil || result.Total != 0 || len(result.Items) != 0 {
		t.Fatalf("public search deleted stack: %+v, %v", result, err)
	}
	result, err = repo.SearchByOwner(ctx, owner.ExternalID, stackDomain.SearchOptions{Query: suffix, SortBy: "name"})
	if err != nil || result.Total != 1 || len(result.Items) != 1 || result.Items[0].UUID != private.UUID {
		t.Fatalf("owner search deleted stack: %+v, %v", result, err)
	}
	count, err = repo.CountPublic(ctx)
	if err != nil || count != baseline {
		t.Fatalf("public count after delete: %d, %v", count, err)
	}
	stats, err = repo.StatsByOwner(ctx, owner.ExternalID)
	if err != nil || stats != (stackDomain.Stats{TotalStacks: 1}) {
		t.Fatalf("stats after delete: %+v, %v", stats, err)
	}
	cancelled, stop := context.WithCancel(ctx)
	stop()
	if _, err := repo.List(cancelled, owner.ExternalID); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled query: %v", err)
	}
	var links, papers int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM public.stack_paper WHERE uuid_stack=$1`, stack.UUID).Scan(&links); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM public.paper WHERE uuid=$1`, paper.UUID).Scan(&papers); err != nil {
		t.Fatal(err)
	}
	if links != 1 || papers != 1 {
		t.Fatalf("delete affected papers: links=%d papers=%d", links, papers)
	}
	t.Run("SQL inputs remain literal values", func(t *testing.T) {
		literal := private
		literal.UUID = failedID
		literal.Name = "Literal ' OR TRUE -- " + suffix
		if err := repo.Create(ctx, literal); err != nil {
			t.Fatal(err)
		}
		got, err := repo.GetByUUID(ctx, literal.UUID)
		if err != nil || got.Name != literal.Name {
			t.Fatalf("SQL-like name changed: %+v, %v", got, err)
		}
		result, err := repo.SearchByOwner(ctx, owner.ExternalID, stackDomain.SearchOptions{Query: literal.Name, SortBy: "name"})
		if err != nil || result.Total != 1 || len(result.Items) != 1 || result.Items[0].UUID != literal.UUID {
			t.Fatalf("SQL-like query was not literal: %+v, %v", result, err)
		}
		result, err = repo.SearchByOwner(ctx, owner.ExternalID, stackDomain.SearchOptions{Query: "' OR TRUE -- missing-" + suffix, SortBy: "name"})
		if err != nil || result.Total != 0 {
			t.Fatalf("SQL-like query broadened search: %+v, %v", result, err)
		}
		items, err := repo.List(ctx, owner.ExternalID+"' OR TRUE --")
		if err != nil || len(items) != 0 {
			t.Fatalf("SQL-like owner bypassed filtering: %+v, %v", items, err)
		}
		if _, err := repo.SearchPublic(ctx, stackDomain.SearchOptions{SortBy: "name; SELECT 1 --"}); !errors.Is(err, stackDomain.ErrInvalidSearch) {
			t.Fatalf("SQL-like sort accepted: %v", err)
		}
		if _, err := repo.GetByUUID(ctx, "' OR TRUE --"); err == nil {
			t.Fatal("SQL-like UUID accepted")
		}
	})
}
