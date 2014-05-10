module Travis
  module Build
    class Script
      module Addons
        class Artifacts
          attr_accessor :script, :config

          def initialize(script, config)
            @script = script
            @config = config
          end

          def after_script
            script.if(want) { run }
          end

          private

          def run
            script.fold('artifacts.0') { install }
            script.fold('artifacts.1') do
              script.cmd(
                "artifacts upload #{options}",
                echo: false,
                assert: false
              )
            end
          end

          def options
            ''
          end
        end
      end
    end
  end
end
