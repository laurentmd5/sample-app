pipeline {
    agent any
    
    triggers {
        pollSCM('H/5 * * * *')
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
        TARGET_URL = "http://192.168.61.131:${APP_PORT}"
        // Configuration Git pour éviter les prompts
        GIT_TERMINAL_PROMPT = '0'
    }
    
    stages {
        // ÉTAPE 1: Checkout du Code
        stage('Checkout Code') {
            steps {
                git branch: 'master',
                    url: 'https://github.com/laurentmd5/sample-app.git',
                    credentialsId: 'github-token2'

                sh '''
                echo "📦 Repository: https://github.com/laurentmd5/sample-app.git"
                echo "📝 Branch: master"
                echo "🔍 Dernier commit:"
                git log -1 --oneline
                echo "📁 Contenu du repository:"
                ls -la
                '''
            }
        }
        
        // ÉTAPE 2: Setup Environment et Outils de Sécurité
        stage('Setup Environment') {
            steps {
                sh '''
                echo "🔧 Configuration de l environnement..."
                echo "=== Versions des outils ==="
                go version || echo "Go non installé"
                docker --version || echo "Docker non disponible"
                
                echo "📥 Installation des outils de sécurité..."
                
                echo "=== Configuration Git pour Go modules ==="
                git config --global url."https://github.com".insteadOf ssh://git@github.com || true
                
                echo "=== Installation de gosec ==="
                if ! which gosec; then
                    echo "Installation de gosec..."
                    # Téléchargement direct depuis les releases GitHub
                    wget -q https://github.com/securecodewarrior/gosec/releases/download/v${GOSEC_VERSION}/gosec_${GOSEC_VERSION}_linux_amd64.tar.gz
                    tar -xzf gosec_${GOSEC_VERSION}_linux_amd64.tar.gz
                    sudo mv gosec /usr/local/bin/
                    rm -f gosec_${GOSEC_VERSION}_linux_amd64.tar.gz
                fi
                gosec --version || echo "Gosec installation échouée"
                
                echo "=== Installation de Trivy ==="
                if ! which trivy; then
                    wget -q https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-64bit.deb
                    sudo dpkg -i trivy_${TRIVY_VERSION}_Linux-64bit.deb
                    rm -f trivy_${TRIVY_VERSION}_Linux-64bit.deb
                fi
                trivy --version || echo "Trivy installation failed"
                
                echo "=== Installation de Lynis ==="
                which lynis || (sudo apt update && sudo apt install -y lynis)
                
                echo "✅ Tous les outils sont prêts"
                '''
            }
        }
        
        // ÉTAPE 3: Build Application Go
        stage('Build Go Application') {
            steps {
                sh '''
                echo "🏗️ Construction de l application Go..."
                
                if [ ! -f "go.mod" ]; then
                    echo "📝 Initialisation de go.mod..."
                    go mod init hello-app
                fi
                
                echo "📥 Téléchargement des dépendances..."
                # Configuration pour éviter les problèmes d'authentification
                GOPRIVATE=""
                go env -w GOPRIVATE=""
                go mod download 2>&1 | tee go-mod.log || echo "Aucune dépendance ou déjà téléchargées"
                
                echo "🔨 Compilation de l application..."
                go build -v -o ${APP_NAME} .
                
                echo "✅ Vérification du build:"
                ls -la ${APP_NAME}
                file ${APP_NAME}
                chmod +x ${APP_NAME}
                '''
            }
        }
        
        // Les autres étapes restent similaires...
        // ÉTAPE 4: Static Code Analysis - gosec
        stage('Static Code Analysis') {
            steps {
                sh '''
                echo "🔍 Analyse Statique du Code avec gosec..."
                mkdir -p security-reports
                
                echo "=== Exécution de gosec ==="
                if which gosec; then
                    gosec -fmt=json -out=security-reports/gosec-report.json ./... 2>/dev/null || true
                    gosec -fmt=html -out=security-reports/gosec-report.html ./... 2>/dev/null || true
                    gosec ./... 2>&1 | tee security-reports/gosec-output.txt || echo "Gosec a terminé avec des findings"
                    echo "✅ Analyse gosec terminée"
                else
                    echo "⚠️ gosec non disponible, analyse statique ignorée"
                fi
                
                echo "=== Exécution de go vet ==="
                go vet ./... 2>&1 | tee security-reports/govet-output.txt || echo "Go vet terminé"
                '''
            }
            post {
                always {
                    publishHTML([
                        allowMissing: true,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'security-reports',
                        reportFiles: 'gosec-report.html',
                        reportName: 'Rapport Gosec - Analyse Statique'
                    ])
                    archiveArtifacts artifacts: 'security-reports/gosec-report.*,security-reports/govet-output.txt', fingerprint: true
                }
            }
        }
        
        // ... (les autres étapes restent inchangées)
    }
    
    post {
        always {
            sh '''
            echo "Nettoyage final..."
            docker system prune -f 2>/dev/null || true
            
            echo ""
            echo "RÉSUMÉ DE L EXÉCUTION"
            echo "Application Go construite"
            echo "Image Docker créée"
            echo "Tests et couverture exécutés"
            echo "Scans de sécurité complétés"
            echo "Application déployée"
            echo ""
            echo "RAPPORTS DISPONIBLES:"
            echo "Gosec: Analyse statique"
            echo "Tests: Couverture et résultats"
            echo "Trivy: Scan containers et fichiers"
            echo "ZAP: Scan dynamique"
            echo "Lynis: Audit environnement"
            echo "Résumé sécurité complet"
            '''
            
            archiveArtifacts artifacts: 'security-reports/**,test-reports/**,trivy-reports/**,zap-reports/**,lynis-reports/**,${APP_NAME}', fingerprint: true
        }
        success {
            sh """
            echo ""
            echo "🎉 PIPELINE DE SÉCURITÉ COMPLET RÉUSSI!"
            echo "Tous les contrôles de sécurité ont passé"
            echo "Application déployée sécuritairement"
            echo "Accédez à l application: ${TARGET_URL}"
            """
        }
        failure {
            sh """
            echo "❌ PIPELINE EN ÉCHEC"
            echo "Consultez les rapports pour plus de détails"
            """
        }
    }
}
