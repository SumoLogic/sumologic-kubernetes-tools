import subprocess

for i in range(10):
    subprocess.Popen(["go", "run", "main.go", "-config", f"config{i+1}.yml"])
