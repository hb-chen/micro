# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: alert/alert.proto

require 'google/protobuf'

Google::Protobuf::DescriptorPool.generated_pool.build do
  add_message "alert.Event" do
    optional :id, :string, 1
    optional :category, :string, 2
    optional :action, :string, 3
    optional :label, :string, 4
    optional :value, :uint64, 5
    map :metadata, :string, :string, 6
  end
  add_message "alert.ReportEventRequest" do
    optional :event, :message, 1, "alert.Event"
  end
  add_message "alert.ReportEventResponse" do
  end
end

module Alert
  Event = Google::Protobuf::DescriptorPool.generated_pool.lookup("alert.Event").msgclass
  ReportEventRequest = Google::Protobuf::DescriptorPool.generated_pool.lookup("alert.ReportEventRequest").msgclass
  ReportEventResponse = Google::Protobuf::DescriptorPool.generated_pool.lookup("alert.ReportEventResponse").msgclass
end
