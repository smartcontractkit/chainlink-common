package store

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/core/store/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	rawURL := os.Getenv("DATABASE_URL")
	url, err := url.Parse(rawURL)
	require.NoError(t, err)

	// connect to DB
	store := Store{}
	t.Run("connect-migrate", func(t *testing.T) {
		require.NoError(t, store.Connect(*url))
	})

	// delete job table at the end
	defer store.DB.Migrator().DropTable(&models.Job{})

	// test create job
	createScenarios := []struct {
		name string
		pass bool
		job  models.Job
		err  string
	}{
		{"success-0", true, models.Job{JobID: "random-test"}, ""},
		{"success-1", true, models.Job{JobID: "another-random-test"}, ""},
		{"duplicate-job", false, models.Job{JobID: "random-test"}, "ERROR: duplicate key value violates unique constraint \"jobs_job_id_key\" (SQLSTATE 23505)"},
		{"missing-job-id", false, models.Job{}, "JobID cannot be blank"},
	}
	for _, s := range createScenarios {
		t.Run(s.name, func(t *testing.T) {
			if s.pass {
				assert.NoError(t, store.CreateJob(&s.job))
				return
			}
			fmt.Println(store.CreateJob(&s.job))
			assert.EqualError(t, store.CreateJob(&s.job), s.err)
		})
	}

	// test LoadJobs
	t.Run("load-jobs-match-create", func(t *testing.T) {
		jobs, err := store.LoadJobs()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(jobs))
		assert.Equal(t, "random-test", jobs[0].JobID)
		assert.Equal(t, "another-random-test", jobs[1].JobID)
	})

	// test delete job
	deleteScenarios := []struct {
		name string
		pass bool
		job  string
		err  string
	}{
		{"succes", true, "random-test", ""},
		{"fail-duplicate", false, "random-test", "record not found"},
		{"fail-empty", false, "", "JobID cannot be blank"},
	}
	for _, s := range deleteScenarios {
		t.Run(s.name, func(t *testing.T) {
			if s.pass {
				assert.NoError(t, store.DeleteJob(s.job))
				return
			}
			assert.EqualError(t, store.DeleteJob(s.job), s.err)
		})
	}

}
