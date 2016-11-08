def run(command)
    puts command
    system command
end

version = `git tag`.lines.last.strip
commit = `git rev-parse HEAD`.strip

run("go build -ldflags \"-w -s -X main.version=#{version} -X main.commit=#{commit}\"")
