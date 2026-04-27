package gengroovy_test

import (
	"testing"

	"github.com/ndscm/theseed/seed/infra/groovy"
	"github.com/ndscm/theseed/seed/infra/groovy/generator/gengroovy"
)

func TestGenerateFromJenkinsDeclarativePipeline(t *testing.T) {
	t.Run("basic_agent_stages_and_steps", func(t *testing.T) {
		source := `pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        sh('make build')
      }
    }
  }
}`
		expected := source
		module, err := groovy.Parse(source)
		if err != nil {
			t.Fatal(err)
		}
		got, err := gengroovy.GenerateIndent(module, "  ")
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Fatalf("GenerateIndent() =\n%q\nexpected\n%q", got, expected)
		}
	})

	t.Run("environment_options_and_post", func(t *testing.T) {
		source := `pipeline {
  agent any
  options {
    timestamps()
    timeout(time: 30, unit: 'MINUTES')
  }
  environment {
    IMAGE = 'theseed/app'
  }
  post {
    always {
      archiveArtifacts('build/**')
    }
    failure {
      mail(to: 'ops@example.com', subject: 'failed')
    }
  }
}`
		expected := source
		module, err := groovy.Parse(source)
		if err != nil {
			t.Fatal(err)
		}
		got, err := gengroovy.GenerateIndent(module, "  ")
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Fatalf("GenerateIndent() =\n%q\nexpected\n%q", got, expected)
		}
	})

	t.Run("array_of_array", func(t *testing.T) {
		source := `pipeline {
agent any
triggers {
GenericTrigger(
causeString: 'Pull Request',
genericHeaderVariables: [[
key: 'x-github-event', defaultValue: '',
]],
genericVariables: [[
key: 'action', value: '$.action', defaultValue: '',
], [
key: 'pull_request_id', value: '$.pull_request.id', defaultValue: '',
]],
regexpFilterText: '$x_github_event $action',
regexpFilterExpression: '^pull_request (opened|reopened|synchronize)$',
)
}
}`
		expected := `pipeline {
  agent any
  triggers {
    GenericTrigger(
      causeString: 'Pull Request',
      genericHeaderVariables: [[
        key: 'x-github-event', defaultValue: '',
      ]],
      genericVariables: [[
        key: 'action', value: '$.action', defaultValue: '',
      ], [
        key: 'pull_request_id', value: '$.pull_request.id', defaultValue: '',
      ]],
      regexpFilterText: '$x_github_event $action',
      regexpFilterExpression: '^pull_request (opened|reopened|synchronize)$',
    )
  }
}`
		module, err := groovy.Parse(source)
		if err != nil {
			t.Fatal(err)
		}
		got, err := gengroovy.GenerateIndent(module, "  ")
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Fatalf("GenerateIndent() =\n%q\nexpected\n%q", got, expected)
		}
	})

	t.Run("seed_devprod_workflow_format", func(t *testing.T) {
		source := `pipeline{agent any
parameters{// repository_url is kept for compatibility with callers that pass it through;
// the Validate stage rejects anything other than the default.
string name:'repository_url',defaultValue:'https://github.com/ndscm/theseed'
string    name:'commit_hash'}
stages{stage('Validate'){
steps{echo    "Repository URL: ${env.repository_url}"
echo "Commit hash: ${env.commit_hash}"


script{if (env.repository_url != 'https://github.com/ndscm/theseed'){
error   "Unexpected repository_url: ${env.repository_url}"
}}
withCredentials([usernamePassword(credentialsId:'ndscm', usernameVariable:'GH_USER', passwordVariable:'GH_TOKEN')]){sh '''#!/usr/bin/env bash
gh api --method POST /repos/ndscm/theseed/statuses/"$commit_hash" \
  -f state=pending \
  -f context=theseed/format \
  -f description='Running format checks'
'''
}}
}
stage('Checkout'){steps{checkout scmGit(
branches:[[name:"${env.commit_hash}"]],
extensions:[
cleanBeforeCheckout(),
cleanAfterCheckout(),
],
userRemoteConfigs:[[url:"${env.repository_url}.git", credentialsId:'ndscm']],
)}}
stage('Format') {steps{sh '''#!/usr/bin/env bash
./format.sh
'''
}}
stage('Test'){
steps{sh '''#!/usr/bin/env bash
if [ -n "$(git status --porcelain)" ]; then
  echo "Worktree is dirty after format:"
  git diff
  exit 1
fi
'''
}}}
post{success{withCredentials([usernamePassword(credentialsId:'ndscm', usernameVariable:'GH_USER', passwordVariable:'GH_TOKEN')]){sh '''#!/usr/bin/env bash
gh api --method POST /repos/ndscm/theseed/statuses/"$commit_hash" \
  -f state=success \
  -f context=theseed/format \
  -f description='Changed files are formatted'
'''
}}


failure {withCredentials([usernamePassword(credentialsId:'ndscm', usernameVariable:'GH_USER', passwordVariable:'GH_TOKEN')]){sh '''#!/usr/bin/env bash
gh api --method POST /repos/ndscm/theseed/statuses/"$commit_hash" \
  -f state=failure \
  -f context=theseed/format \
  -f description='Format pipeline failed'
'''
}}}
}`
		expected := `pipeline {
  agent any
  parameters {
    // repository_url is kept for compatibility with callers that pass it through;
    // the Validate stage rejects anything other than the default.
    string name: 'repository_url', defaultValue: 'https://github.com/ndscm/theseed'
    string name: 'commit_hash'
  }
  stages {
    stage('Validate') {
      steps {
        echo "Repository URL: ${env.repository_url}"
        echo "Commit hash: ${env.commit_hash}"
        script {
          if (env.repository_url != 'https://github.com/ndscm/theseed') {
            error "Unexpected repository_url: ${env.repository_url}"
          }
        }
        withCredentials([usernamePassword(credentialsId: 'ndscm', usernameVariable: 'GH_USER', passwordVariable: 'GH_TOKEN')]) {
          sh '''#!/usr/bin/env bash
gh api --method POST /repos/ndscm/theseed/statuses/"$commit_hash" \
  -f state=pending \
  -f context=theseed/format \
  -f description='Running format checks'
'''
        }
      }
    }
    stage('Checkout') {
      steps {
        checkout scmGit(
          branches: [[name: "${env.commit_hash}"]],
          extensions: [
            cleanBeforeCheckout(),
            cleanAfterCheckout(),
          ],
          userRemoteConfigs: [[url: "${env.repository_url}.git", credentialsId: 'ndscm']],
        )
      }
    }
    stage('Format') {
      steps {
        sh '''#!/usr/bin/env bash
./format.sh
'''
      }
    }
    stage('Test') {
      steps {
        sh '''#!/usr/bin/env bash
if [ -n "$(git status --porcelain)" ]; then
  echo "Worktree is dirty after format:"
  git diff
  exit 1
fi
'''
      }
    }
  }
  post {
    success {
      withCredentials([usernamePassword(credentialsId: 'ndscm', usernameVariable: 'GH_USER', passwordVariable: 'GH_TOKEN')]) {
        sh '''#!/usr/bin/env bash
gh api --method POST /repos/ndscm/theseed/statuses/"$commit_hash" \
  -f state=success \
  -f context=theseed/format \
  -f description='Changed files are formatted'
'''
      }
    }
    failure {
      withCredentials([usernamePassword(credentialsId: 'ndscm', usernameVariable: 'GH_USER', passwordVariable: 'GH_TOKEN')]) {
        sh '''#!/usr/bin/env bash
gh api --method POST /repos/ndscm/theseed/statuses/"$commit_hash" \
  -f state=failure \
  -f context=theseed/format \
  -f description='Format pipeline failed'
'''
      }
    }
  }
}`
		module, err := groovy.Parse(source)
		if err != nil {
			t.Fatal(err)
		}
		got, err := gengroovy.GenerateIndent(module, "  ")
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Fatalf("GenerateIndent() =\n%q\nexpected\n%q", got, expected)
		}
	})
}
