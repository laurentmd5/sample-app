pipeline {
  agent any
  stages {
    stage('Checkout Code') {
      steps {
        git(branch: 'master', url: 'https://github.com/laurentmd5/sample-app.git', credentialsId: '732f6a82-d7a3-4316-860d-382960b3d9da')
        sh '''
          echo "📦 Repository: https://github.com/laurentmd5/sample-app.git"
          echo "📝 Branch: master"
          ls -la
        '''
      }
    }

    stage('Setup Environment') {
      steps {
        sh '''
          set -e
          echo "🔧 Configuration des outils de sécurité..."

          # --- Installation de GoSec ---
          if ! command -v gosec >/dev/null 2>&1; then
            echo "📥 Installation de GoSec v${GOSEC_VERSION}..."
            wget -q https://github.com/securego/gosec/releases/download/v${GOSEC_VERSION}/gosec_${GOSEC_VERSION}_linux_amd64.tar.gz               && tar -xzf gosec_${GOSEC_VERSION}_linux_amd64.tar.gz               && mv gosec /usr/local/bin/               && echo "✅ GoSec installé avec succès."
          else
            echo "🔹 GoSec déjà présent : $(gosec --version)"
          fi

          # --- Installation de Trivy ---
          if ! command -v trivy >/dev/null 2>&1; then
            echo "📥 Installation de Trivy v${TRIVY_VERSION}..."
            wget -q https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-64bit.deb               && dpkg -i trivy_${TRIVY_VERSION}_Linux-64bit.deb               && echo "✅ Trivy installé avec succès."
          else
            echo "🔹 Trivy déjà présent : $(trivy --version | head -n 1)"
          fi

          # --- Installation de Lynis ---
          if ! command -v lynis >/dev/null 2>&1; then
            echo "📥 Installation de Lynis..."
            apt-get update -y >/dev/null 2>&1               && apt-get install -y lynis >/dev/null 2>&1               && echo "✅ Lynis installé avec succès."
          else
            echo "🔹 Lynis déjà présent : $(lynis show version)"
          fi

          echo "✅ Environnement de sécurité configuré avec succès."
        '''
      }
    }

    stage('Build Go Application') {
      steps {
        sh '''
          echo "🏗️ Construction de l’application Go..."
          go mod tidy || true
          go build -o ${APP_NAME} .
          file ${APP_NAME}
        '''
      }
    }

    stage('Static Code Analysis') {
      post {
        always {
          publishHTML(reportDir: 'security-reports', reportFiles: 'gosec-report.html', reportName: 'Analyse Statique (GoSec)', allowMissing: true, keepAll: true, alwaysLinkToLastBuild: true)
        }

      }
      steps {
        sh '''
          echo "🔍 Analyse Statique du Code avec GoSec..."
          mkdir -p security-reports
          gosec -fmt=html -out=security-reports/gosec-report.html ./... || true
        '''
      }
    }

    stage('Build Docker Image') {
      steps {
        sh '''
          echo "🐳 Construction de l’image Docker..."
          docker build -t ${DOCKER_REGISTRY}/${APP_NAME}:latest .
        '''
      }
    }

    stage('Container Scan - Trivy') {
      post {
        always {
          publishHTML(reportDir: 'trivy-reports', reportFiles: 'container-scan.html', reportName: 'Analyse Conteneur (Trivy)', allowMissing: true, keepAll: true, alwaysLinkToLastBuild: true)
        }

      }
      steps {
        sh '''
          echo "🧩 Scan de vulnérabilités du conteneur..."
          mkdir -p trivy-reports
          trivy image --format html -o trivy-reports/container-scan.html ${DOCKER_REGISTRY}/${APP_NAME}:latest || true
        '''
      }
    }

    stage('Deploy to Ubuntu via SSH') {
      steps {
        script {
          withCredentials([sshUserPrivateKey(
            credentialsId: "${SSH_CREDENTIALS_ID}",
            usernameVariable: 'SSH_USER',
            keyFileVariable: 'SSH_KEY'
          )]) {
            sh """
            ssh -i \${SSH_KEY} -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
            docker stop ${APP_NAME} 2>/dev/null || true
            docker rm ${APP_NAME} 2>/dev/null || true
            docker run -d --name ${APP_NAME} -p ${APP_PORT}:${APP_PORT} ${DOCKER_REGISTRY}/${APP_NAME}:latest
            "
            """
          }
        }

      }
    }

    stage('Dynamic Analysis - OWASP ZAP') {
      post {
        always {
          publishHTML(reportDir: 'zap-reports', reportFiles: "${ZAP_REPORT}", reportName: 'Analyse Dynamique OWASP ZAP', allowMissing: true, keepAll: true, alwaysLinkToLastBuild: true)
          archiveArtifacts(artifacts: 'zap-reports/*', fingerprint: true)
        }

      }
      steps {
        sh '''
          echo "⚡ Lancement du scan OWASP ZAP via Docker..."
          mkdir -p zap-reports
          docker run --rm -u root             -v $(pwd)/zap-reports:/zap/wrk             -t owasp/zap2docker-stable:${ZAP_VERSION}             zap-baseline.py -t ${TARGET_URL} -r ${ZAP_REPORT} -J zap-report.json -z "-config api.key=${ZAP_API_KEY}"

          echo "✅ Rapport OWASP ZAP généré dans zap-reports/${ZAP_REPORT}"
        '''
      }
    }

    stage('Final Security Summary') {
      post {
        always {
          publishHTML(reportDir: 'security-summary', reportFiles: 'summary.html', reportName: 'Résumé Global Sécurité', allowMissing: true, keepAll: true, alwaysLinkToLastBuild: true)
        }

      }
      steps {
        sh '''
          echo "📊 Compilation du rapport de sécurité global..."
          mkdir -p security-summary
          cat > security-summary/summary.html <<EOF
          <html><body>
          <h2>Rapport Global Sécurité - Build ${BUILD_NUMBER}</h2>
          <ul>
            <li>Analyse Statique (GoSec)</li>
            <li>Analyse Conteneur (Trivy)</li>
            <li>Analyse Dynamique (OWASP ZAP)</li>
          </ul>
          <p>Status Global : ${currentBuild.currentResult}</p>
          </body></html>
          EOF
        '''
      }
    }

  }
  environment {
    APP_NAME = 'hello-app'
    APP_PORT = '8090'
    DOCKER_REGISTRY = 'laurentmd5'
    DEPLOY_SERVER = 'devops@localhost'
    SSH_CREDENTIALS_ID = 'ubuntu-server-ssh'
    TRIVY_VERSION = '0.49.1'
    GOSEC_VERSION = '2.19.0'
    ZAP_VERSION = '2.14.0'
    TARGET_URL = 'http://192.168.61.131:8090'
    ZAP_REPORT = 'zap-report.html'
    ZAP_API_KEY = '12345'
    GIT_TERMINAL_PROMPT = '0'
  }
  post {
    success {
      echo '🎉 Pipeline complet exécuté avec succès !'
    }

    failure {
      echo '❌ Une étape du pipeline a échoué. Vérifiez les rapports HTML dans Jenkins.'
    }

  }
}
